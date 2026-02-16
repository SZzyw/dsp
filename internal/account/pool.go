package account

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"ds2api/internal/config"
)

type Pool struct {
	store                 *config.Store
	mu                    sync.Mutex
	queue                 []string
	inUse                 map[string]int
	maxInflightPerAccount int
	recommendedConcurrency int
}

func NewPool(store *config.Store) *Pool {
	p := &Pool{
		store:                 store,
		inUse:                 map[string]int{},
		maxInflightPerAccount: maxInflightFromEnv(),
	}
	p.Reset()
	return p
}

func (p *Pool) Reset() {
	accounts := p.store.Accounts()
	sort.SliceStable(accounts, func(i, j int) bool {
		iHas := accounts[i].Token != ""
		jHas := accounts[j].Token != ""
		if iHas == jHas {
			return i < j
		}
		return iHas
	})
	ids := make([]string, 0, len(accounts))
	for _, a := range accounts {
		id := a.Identifier()
		if id != "" {
			ids = append(ids, id)
		}
	}
	recommended := defaultRecommendedConcurrency(len(ids), p.maxInflightPerAccount)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queue = ids
	p.inUse = map[string]int{}
	p.recommendedConcurrency = recommended
	config.Logger.Info(
		"[init_account_queue] initialized",
		"total", len(ids),
		"max_inflight_per_account", p.maxInflightPerAccount,
		"recommended_concurrency", p.recommendedConcurrency,
	)
}

func (p *Pool) Acquire(target string, exclude map[string]bool) (config.Account, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if exclude == nil {
		exclude = map[string]bool{}
	}
	if target != "" {
		if exclude[target] || p.inUse[target] >= p.maxInflightPerAccount {
			return config.Account{}, false
		}
		acc, ok := p.store.FindAccount(target)
		if !ok {
			return config.Account{}, false
		}
		p.inUse[target]++
		p.bumpQueue(target)
		return acc, true
	}

	if acc, ok := p.tryAcquire(exclude, true); ok {
		return acc, true
	}
	if acc, ok := p.tryAcquire(exclude, false); ok {
		return acc, true
	}
	return config.Account{}, false
}

func (p *Pool) tryAcquire(exclude map[string]bool, requireToken bool) (config.Account, bool) {
	for i := 0; i < len(p.queue); i++ {
		id := p.queue[i]
		if exclude[id] || p.inUse[id] >= p.maxInflightPerAccount {
			continue
		}
		acc, ok := p.store.FindAccount(id)
		if !ok {
			continue
		}
		if requireToken && acc.Token == "" {
			continue
		}
		p.inUse[id]++
		p.bumpQueue(id)
		return acc, true
	}
	return config.Account{}, false
}

func (p *Pool) bumpQueue(accountID string) {
	for i, id := range p.queue {
		if id != accountID {
			continue
		}
		p.queue = append(p.queue[:i], p.queue[i+1:]...)
		p.queue = append(p.queue, accountID)
		return
	}
}

func (p *Pool) Release(accountID string) {
	if accountID == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	count := p.inUse[accountID]
	if count <= 0 {
		return
	}
	if count == 1 {
		delete(p.inUse, accountID)
		return
	}
	p.inUse[accountID] = count - 1
}

func (p *Pool) Status() map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	available := make([]string, 0, len(p.queue))
	inUseAccounts := make([]string, 0, len(p.inUse))
	inUseSlots := 0
	for _, id := range p.queue {
		if p.inUse[id] < p.maxInflightPerAccount {
			available = append(available, id)
		}
	}
	for id, count := range p.inUse {
		if count > 0 {
			inUseAccounts = append(inUseAccounts, id)
			inUseSlots += count
		}
	}
	sort.Strings(inUseAccounts)
	return map[string]any{
		"available":                len(available),
		"in_use":                   inUseSlots,
		"total":                    len(p.store.Accounts()),
		"available_accounts":       available,
		"in_use_accounts":          inUseAccounts,
		"max_inflight_per_account": p.maxInflightPerAccount,
		"recommended_concurrency":  p.recommendedConcurrency,
	}
}

func maxInflightFromEnv() int {
	for _, key := range []string{"DS2API_ACCOUNT_MAX_INFLIGHT", "DS2API_ACCOUNT_CONCURRENCY"} {
		raw := strings.TrimSpace(os.Getenv(key))
		if raw == "" {
			continue
		}
		n, err := strconv.Atoi(raw)
		if err == nil && n > 0 {
			return n
		}
	}
	return 2
}

func defaultRecommendedConcurrency(accountCount, maxInflightPerAccount int) int {
	if accountCount <= 0 {
		return 0
	}
	if maxInflightPerAccount <= 0 {
		maxInflightPerAccount = 2
	}
	return accountCount * maxInflightPerAccount
}
