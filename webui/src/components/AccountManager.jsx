import { useState, useEffect } from 'react'
import {
    Plus,
    Trash2,
    RefreshCw,
    CheckCircle2,
    AlertCircle,
    Search,
    Play,
    MoreHorizontal,
    X,
    Server,
    ShieldCheck
} from 'lucide-react'
import clsx from 'clsx'

export default function AccountManager({ config, onRefresh, onMessage, authFetch }) {
    const [showAddKey, setShowAddKey] = useState(false)
    const [showAddAccount, setShowAddAccount] = useState(false)
    const [newKey, setNewKey] = useState('')
    const [newAccount, setNewAccount] = useState({ email: '', mobile: '', password: '' })
    const [loading, setLoading] = useState(false)
    const [validating, setValidating] = useState({})
    const [validatingAll, setValidatingAll] = useState(false)
    const [testing, setTesting] = useState({})
    const [testingAll, setTestingAll] = useState(false)
    const [batchProgress, setBatchProgress] = useState({ current: 0, total: 0, results: [] })
    const [queueStatus, setQueueStatus] = useState(null)

    const apiFetch = authFetch || fetch

    const fetchQueueStatus = async () => {
        try {
            const res = await apiFetch('/admin/queue/status')
            if (res.ok) {
                const data = await res.json()
                setQueueStatus(data)
            }
        } catch (e) {
            console.error('Failed to fetch queue status:', e)
        }
    }

    useEffect(() => {
        fetchQueueStatus()
        const interval = setInterval(fetchQueueStatus, 5000)
        return () => clearInterval(interval)
    }, [])

    const addKey = async () => {
        if (!newKey.trim()) return
        setLoading(true)
        try {
            const res = await apiFetch('/admin/keys', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ key: newKey.trim() }),
            })
            if (res.ok) {
                onMessage('success', 'API Key added successfully')
                setNewKey('')
                setShowAddKey(false)
                onRefresh()
            } else {
                const data = await res.json()
                onMessage('error', data.detail || 'Failed to add')
            }
        } catch (e) {
            onMessage('error', 'Network error')
        } finally {
            setLoading(false)
        }
    }

    const deleteKey = async (key) => {
        if (!confirm('Are you sure you want to delete this API Key?')) return
        try {
            const res = await apiFetch(`/admin/keys/${encodeURIComponent(key)}`, { method: 'DELETE' })
            if (res.ok) {
                onMessage('success', 'Deleted successfully')
                onRefresh()
            } else {
                onMessage('error', 'Delete failed')
            }
        } catch (e) {
            onMessage('error', 'Network error')
        }
    }

    const addAccount = async () => {
        if (!newAccount.password || (!newAccount.email && !newAccount.mobile)) {
            onMessage('error', 'Password and Email/Mobile are required')
            return
        }
        setLoading(true)
        try {
            const res = await apiFetch('/admin/accounts', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(newAccount),
            })
            if (res.ok) {
                onMessage('success', 'Account added successfully')
                setNewAccount({ email: '', mobile: '', password: '' })
                setShowAddAccount(false)
                onRefresh()
            } else {
                const data = await res.json()
                onMessage('error', data.detail || 'Failed to add')
            }
        } catch (e) {
            onMessage('error', 'Network error')
        } finally {
            setLoading(false)
        }
    }

    const deleteAccount = async (id) => {
        if (!confirm('Are you sure you want to delete this account?')) return
        try {
            const res = await apiFetch(`/admin/accounts/${encodeURIComponent(id)}`, { method: 'DELETE' })
            if (res.ok) {
                onMessage('success', 'Deleted successfully')
                onRefresh()
            } else {
                onMessage('error', 'Delete failed')
            }
        } catch (e) {
            onMessage('error', 'Network error')
        }
    }

    const validateAccount = async (identifier) => {
        setValidating(prev => ({ ...prev, [identifier]: true }))
        try {
            const res = await apiFetch('/admin/accounts/validate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ identifier }),
            })
            const data = await res.json()
            onMessage(data.valid ? 'success' : 'error', `${identifier}: ${data.message}`)
            onRefresh()
        } catch (e) {
            onMessage('error', 'Validation failed: ' + e.message)
        } finally {
            setValidating(prev => ({ ...prev, [identifier]: false }))
        }
    }

    const validateAllAccounts = async () => {
        if (!confirm('Validate ALL accounts? This might take a while.')) return
        const accounts = config.accounts || []
        if (accounts.length === 0) return

        setValidatingAll(true)
        setBatchProgress({ current: 0, total: accounts.length, results: [] })

        let validCount = 0
        const results = []

        for (let i = 0; i < accounts.length; i++) {
            const acc = accounts[i]
            const id = acc.email || acc.mobile

            try {
                const res = await apiFetch('/admin/accounts/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ identifier: id }),
                })
                const data = await res.json()
                results.push({ id, success: data.valid, message: data.message })
                if (data.valid) validCount++
            } catch (e) {
                results.push({ id, success: false, message: e.message })
            }

            setBatchProgress({ current: i + 1, total: accounts.length, results: [...results] })
        }

        onMessage('success', `Completed: ${validCount}/${accounts.length} valid`)
        onRefresh()
        setValidatingAll(false)
    }

    const testAccount = async (identifier) => {
        setTesting(prev => ({ ...prev, [identifier]: true }))
        try {
            const res = await apiFetch('/admin/accounts/test', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ identifier }),
            })
            const data = await res.json()
            onMessage(data.success ? 'success' : 'error', `${identifier}: ${data.success ? `Success (${data.response_time}ms)` : data.message}`)
            onRefresh()
        } catch (e) {
            onMessage('error', 'Test failed: ' + e.message)
        } finally {
            setTesting(prev => ({ ...prev, [identifier]: false }))
        }
    }

    const testAllAccounts = async () => {
        if (!confirm('Test API connectivity for ALL accounts?')) return
        const accounts = config.accounts || []
        if (accounts.length === 0) return

        setTestingAll(true)
        setBatchProgress({ current: 0, total: accounts.length, results: [] })

        let successCount = 0
        const results = []

        for (let i = 0; i < accounts.length; i++) {
            const acc = accounts[i]
            const id = acc.email || acc.mobile

            try {
                const res = await apiFetch('/admin/accounts/test', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ identifier: id }),
                })
                const data = await res.json()
                results.push({ id, success: data.success, message: data.message, time: data.response_time })
                if (data.success) successCount++
            } catch (e) {
                results.push({ id, success: false, message: e.message })
            }

            setBatchProgress({ current: i + 1, total: accounts.length, results: [...results] })
        }

        onMessage('success', `Completed: ${successCount}/${accounts.length} available`)
        onRefresh()
        setTestingAll(false)
    }

    return (
        <div className="space-y-6">
            {/* Queue Status */}
            {queueStatus && (
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="bg-card border border-border rounded-xl p-3 lg:p-4 flex items-center justify-between shadow-sm">
                        <div className="flex items-center gap-2 lg:gap-3">
                            <div className="p-1.5 lg:p-2 bg-emerald-500/10 text-emerald-500 rounded-lg">
                                <CheckCircle2 className="w-4 h-4 lg:w-5 h-5" />
                            </div>
                            <div>
                                <p className="text-[10px] lg:text-sm font-medium text-muted-foreground uppercase tracking-wider lg:capitalize lg:tracking-normal">可用账号</p>
                                <p className="text-xl lg:text-2xl font-bold">{queueStatus.available}</p>
                            </div>
                        </div>
                    </div>
                    <div className="bg-card border border-border rounded-xl p-3 lg:p-4 flex items-center justify-between shadow-sm">
                        <div className="flex items-center gap-2 lg:gap-3">
                            <div className="p-1.5 lg:p-2 bg-amber-500/10 text-amber-500 rounded-lg">
                                <Server className="w-4 h-4 lg:w-5 h-5" />
                            </div>
                            <div>
                                <p className="text-[10px] lg:text-sm font-medium text-muted-foreground uppercase tracking-wider lg:capitalize lg:tracking-normal">使用中</p>
                                <p className="text-xl lg:text-2xl font-bold">{queueStatus.in_use}</p>
                            </div>
                        </div>
                    </div>
                    <div className="bg-card border border-border rounded-xl p-3 lg:p-4 flex items-center justify-between shadow-sm">
                        <div className="flex items-center gap-2 lg:gap-3">
                            <div className="p-1.5 lg:p-2 bg-primary/10 text-primary rounded-lg">
                                <ShieldCheck className="w-4 h-4 lg:w-5 h-5" />
                            </div>
                            <div>
                                <p className="text-[10px] lg:text-sm font-medium text-muted-foreground uppercase tracking-wider lg:capitalize lg:tracking-normal">账号总数</p>
                                <p className="text-xl lg:text-2xl font-bold">{queueStatus.total}</p>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* API Keys Section */}
            <div className="bg-card border border-border rounded-xl overflow-hidden shadow-sm">
                <div className="p-6 border-b border-border flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div>
                        <h2 className="text-lg font-semibold">API 密钥</h2>
                        <p className="text-sm text-muted-foreground">管理 API 访问密钥池</p>
                    </div>
                    <button
                        onClick={() => setShowAddKey(true)}
                        className="btn btn-primary flex items-center gap-2"
                    >
                        <Plus className="w-4 h-4" />
                        新增密钥
                    </button>
                </div>

                <div className="divide-y divide-border">
                    {config.keys?.length > 0 ? (
                        config.keys.map((key, i) => (
                            <div key={i} className="p-4 flex items-center justify-between hover:bg-muted/50 transition-colors group">
                                <div className="font-mono text-sm bg-muted/50 px-3 py-1 rounded inline-block">
                                    {key.slice(0, 16)}****
                                </div>
                                <button
                                    onClick={() => deleteKey(key)}
                                    className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-md transition-colors opacity-0 group-hover:opacity-100"
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            </div>
                        ))
                    ) : (
                        <div className="p-8 text-center text-muted-foreground">未找到 API 密钥</div>
                    )}
                </div>
            </div>

            {/* Accounts Section */}
            <div className="bg-card border border-border rounded-xl overflow-hidden shadow-sm">
                <div className="p-6 border-b border-border flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div>
                        <h2 className="text-lg font-semibold">DeepSeek 账号</h2>
                        <p className="text-sm text-muted-foreground">管理 DeepSeek 账号池</p>
                    </div>
                    <div className="flex flex-wrap gap-2">
                        <button
                            onClick={testAllAccounts}
                            disabled={testingAll || validatingAll || !config.accounts?.length}
                            className="btn btn-secondary text-xs"
                        >
                            {testingAll ? <span className="animate-spin mr-2">⟳</span> : <Play className="w-3 h-3 mr-2" />}
                            测试全部
                        </button>
                        <button
                            onClick={validateAllAccounts}
                            disabled={validatingAll || testingAll || !config.accounts?.length}
                            className="btn btn-secondary text-xs"
                        >
                            {validatingAll ? <span className="animate-spin mr-2">⟳</span> : <CheckCircle2 className="w-3 h-3 mr-2" />}
                            校验全部
                        </button>
                        <button
                            onClick={() => setShowAddAccount(true)}
                            className="btn btn-primary flex items-center gap-2"
                        >
                            <Plus className="w-4 h-4" />
                            添加账号
                        </button>
                    </div>
                </div>

                {/* Batch Progress */}
                {(testingAll || validatingAll) && batchProgress.total > 0 && (
                    <div className="p-4 border-b border-border bg-muted/30">
                        <div className="flex items-center justify-between text-sm mb-2">
                            <span className="font-medium">{testingAll ? '正在测试所有账号...' : '正在校验所有账号...'}</span>
                            <span className="text-muted-foreground">{batchProgress.current} / {batchProgress.total}</span>
                        </div>
                        <div className="w-full bg-muted rounded-full h-2 overflow-hidden mb-4">
                            <div
                                className="bg-primary h-full transition-all duration-300"
                                style={{ width: `${(batchProgress.current / batchProgress.total) * 100}%` }}
                            />
                        </div>
                        {batchProgress.results.length > 0 && (
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-2 max-h-32 overflow-y-auto custom-scrollbar">
                                {batchProgress.results.map((r, i) => (
                                    <div key={i} className={clsx(
                                        "text-xs px-2 py-1 rounded border truncate",
                                        r.success ? "bg-emerald-500/10 border-emerald-500/20 text-emerald-500" : "bg-destructive/10 border-destructive/20 text-destructive"
                                    )}>
                                        {r.success ? '✓' : '✗'} {r.id}
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                <div className="divide-y divide-border">
                    {config.accounts?.length > 0 ? (
                        config.accounts.map((acc, i) => {
                            const id = acc.email || acc.mobile
                            return (
                                <div key={i} className="p-4 flex flex-col md:flex-row md:items-center justify-between gap-4 hover:bg-muted/50 transition-colors">
                                    <div className="flex items-center gap-3 min-w-0">
                                        <div className={clsx(
                                            "w-2 h-2 rounded-full shrink-0",
                                            acc.has_token ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" : "bg-amber-500"
                                        )} />
                                        <div className="min-w-0">
                                            <div className="font-medium truncate">{id}</div>
                                            <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                                                <span>{acc.has_token ? '已建立会话' : '需重新登录'}</span>
                                                {acc.token_preview && (
                                                    <span className="font-mono bg-muted px-1.5 py-0.5 rounded text-[10px]">
                                                        {acc.token_preview}
                                                    </span>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-2 self-start lg:self-auto ml-5 lg:ml-0">
                                        <button
                                            onClick={() => testAccount(id)}
                                            disabled={testing[id]}
                                            className="px-2 lg:px-3 py-1 lg:py-1.5 text-[10px] lg:text-xs font-medium border border-border rounded-md hover:bg-secondary transition-colors disabled:opacity-50"
                                        >
                                            {testing[id] ? '正在测试...' : '测试'}
                                        </button>
                                        <button
                                            onClick={() => validateAccount(id)}
                                            disabled={validating[id]}
                                            className="px-2 lg:px-3 py-1 lg:py-1.5 text-[10px] lg:text-xs font-medium border border-border rounded-md hover:bg-secondary transition-colors disabled:opacity-50"
                                        >
                                            {validating[id] ? '正在校验...' : '校验'}
                                        </button>
                                        <button
                                            onClick={() => deleteAccount(id)}
                                            className="p-1 lg:p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-md transition-colors"
                                        >
                                            <Trash2 className="w-3.5 h-3.5 lg:w-4 h-4" />
                                        </button>
                                    </div>
                                </div>
                            )
                        })
                    ) : (
                        <div className="p-8 text-center text-muted-foreground">未找到任何账号</div>
                    )}
                </div>
            </div>

            {/* Modals */}
            {showAddKey && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 animate-in fade-in">
                    <div className="bg-card w-full max-w-md rounded-xl border border-border shadow-2xl overflow-hidden animate-in zoom-in-95">
                        <div className="p-4 border-b border-border flex justify-between items-center">
                            <h3 className="font-semibold">添加 API 密钥</h3>
                            <button onClick={() => setShowAddKey(false)} className="text-muted-foreground hover:text-foreground">
                                <X className="w-5 h-5" />
                            </button>
                        </div>
                        <div className="p-6 space-y-4">
                            <div>
                                <label className="block text-sm font-medium mb-1.5">新密钥值</label>
                                <input
                                    type="text"
                                    className="input-field"
                                    placeholder="输入自定义 API 密钥"
                                    value={newKey}
                                    onChange={e => setNewKey(e.target.value)}
                                    autoFocus
                                />
                            </div>
                            <div className="flex justify-end gap-2 pt-2">
                                <button onClick={() => setShowAddKey(false)} className="btn btn-secondary">取消</button>
                                <button onClick={addKey} disabled={loading} className="btn btn-primary">
                                    {loading ? '添加中...' : '添加'}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {showAddAccount && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 animate-in fade-in">
                    <div className="bg-card w-full max-w-md rounded-xl border border-border shadow-2xl overflow-hidden animate-in zoom-in-95">
                        <div className="p-4 border-b border-border flex justify-between items-center">
                            <h3 className="font-semibold">添加 DeepSeek 账号</h3>
                            <button onClick={() => setShowAddAccount(false)} className="text-muted-foreground hover:text-foreground">
                                <X className="w-5 h-5" />
                            </button>
                        </div>
                        <div className="p-6 space-y-4">
                            <div>
                                <label className="block text-sm font-medium mb-1.5">邮箱 (可选)</label>
                                <input
                                    type="email"
                                    className="input-field"
                                    placeholder="user@example.com"
                                    value={newAccount.email}
                                    onChange={e => setNewAccount({ ...newAccount, email: e.target.value })}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1.5">手机号 (可选)</label>
                                <input
                                    type="text"
                                    className="input-field"
                                    placeholder="+86..."
                                    value={newAccount.mobile}
                                    onChange={e => setNewAccount({ ...newAccount, mobile: e.target.value })}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1.5">密码 <span className="text-destructive">*</span></label>
                                <input
                                    type="password"
                                    className="input-field"
                                    placeholder="账号密码"
                                    value={newAccount.password}
                                    onChange={e => setNewAccount({ ...newAccount, password: e.target.value })}
                                />
                            </div>
                            <div className="flex justify-end gap-2 pt-2">
                                <button onClick={() => setShowAddAccount(false)} className="btn btn-secondary">取消</button>
                                <button onClick={addAccount} disabled={loading} className="btn btn-primary">
                                    {loading ? '添加中...' : '添加账号'}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
