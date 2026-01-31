import { useState, useRef } from 'react'
import {
    Send,
    Square,
    MessageSquare,
    Cpu,
    Search as SearchIcon,
    Sparkles,
    Bot,
    User,
    Loader2,
    CheckCircle2,
    AlertCircle,
    ChevronDown,
    ShieldCheck
} from 'lucide-react'
import clsx from 'clsx'

const MODELS = [
    { id: 'deepseek-chat', name: 'DeepSeek Chat', icon: MessageSquare, desc: '通用聊天模型' },
    { id: 'deepseek-reasoner', name: 'DeepSeek Reasoner', icon: Cpu, desc: '推理能力增强模型' },
    // Removed search models as they might be deprecated or identical to chat with search tool
]

export default function ApiTester({ config, onMessage, authFetch }) {
    const [model, setModel] = useState('deepseek-chat')
    const [message, setMessage] = useState('Hello, please introduce yourself in one sentence.')
    const [apiKey, setApiKey] = useState('')
    const [selectedAccount, setSelectedAccount] = useState('')
    const [response, setResponse] = useState(null)
    const [loading, setLoading] = useState(false)
    const [streamingContent, setStreamingContent] = useState('')
    const [streamingThinking, setStreamingThinking] = useState('')
    const [isStreaming, setIsStreaming] = useState(false)
    const abortControllerRef = useRef(null)

    const [sidebarOpen, setSidebarOpen] = useState(false)
    const [configExpanded, setConfigExpanded] = useState(false)

    const apiFetch = authFetch || fetch
    const accounts = config.accounts || []

    const stopGeneration = () => {
        if (abortControllerRef.current) {
            abortControllerRef.current.abort()
            abortControllerRef.current = null
        }
        setLoading(false)
        setIsStreaming(false)
    }

    const directTest = async () => {
        if (loading) return

        setLoading(true)
        setIsStreaming(true)
        setResponse(null)
        setStreamingContent('')
        setStreamingThinking('')

        abortControllerRef.current = new AbortController()

        try {
            const key = apiKey || (config.keys?.[0] || '')
            if (!key) {
                onMessage('error', '请提供 API 密钥')
                setLoading(false)
                setIsStreaming(false)
                return
            }

            const res = await fetch('/v1/chat/completions', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${key}`,
                },
                body: JSON.stringify({
                    model,
                    messages: [{ role: 'user', content: message }],
                    stream: true,
                }),
                signal: abortControllerRef.current.signal,
            })

            if (!res.ok) {
                const data = await res.json()
                setResponse({ success: false, error: data.error?.message || '请求失败' })
                onMessage('error', data.error?.message || '请求失败')
                setLoading(false)
                setIsStreaming(false)
                return
            }

            setResponse({ success: true, status_code: res.status })

            const reader = res.body.getReader()
            const decoder = new TextDecoder()
            let buffer = ''

            while (true) {
                const { done, value } = await reader.read()
                if (done) break

                buffer += decoder.decode(value, { stream: true })
                const lines = buffer.split('\n')
                buffer = lines.pop() || ''

                for (const line of lines) {
                    const trimmed = line.trim()
                    if (!trimmed || !trimmed.startsWith('data: ')) continue

                    const dataStr = trimmed.slice(6)
                    if (dataStr === '[DONE]') continue

                    try {
                        const json = JSON.parse(dataStr)
                        const choice = json.choices?.[0]
                        if (choice?.delta) {
                            const delta = choice.delta
                            if (delta.reasoning_content) {
                                setStreamingThinking(prev => prev + delta.reasoning_content)
                            }
                            if (delta.content) {
                                setStreamingContent(prev => prev + delta.content)
                            }
                        }
                    } catch (e) {
                        console.error('Invalid JSON hunk:', dataStr, e)
                    }
                }
            }
        } catch (e) {
            if (e.name === 'AbortError') {
                onMessage('info', '已停止生成')
            } else {
                onMessage('error', 'Network error: ' + e.message)
                setResponse({ error: e.message, success: false })
            }
        } finally {
            setLoading(false)
            setIsStreaming(false)
            abortControllerRef.current = null
        }
    }

    const sendTest = async () => {
        if (selectedAccount) {
            setLoading(true)
            setResponse(null)
            try {
                const res = await apiFetch('/admin/accounts/test', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        identifier: selectedAccount,
                        model,
                        message,
                    }),
                })
                const data = await res.json()
                setResponse({
                    success: data.success,
                    status_code: res.status,
                    response: data,
                    account: selectedAccount,
                })
                if (data.success) {
                    onMessage('success', `${selectedAccount}: Test Success (${data.response_time}ms)`)
                } else {
                    onMessage('error', `${selectedAccount}: ${data.message}`)
                }
            } catch (e) {
                onMessage('error', 'Network error: ' + e.message)
                setResponse({ error: e.message })
            } finally {
                setLoading(false)
            }
            return
        }

        directTest()
    }

    return (
        <div className="flex flex-col lg:grid lg:grid-cols-3 gap-4 lg:gap-6 h-[calc(100vh-140px)] lg:h-[calc(100vh-140px)]">
            {/* Configuration Panel */}
            <div className={clsx(
                "lg:col-span-1 flex flex-col transition-all duration-300 ease-in-out",
                configExpanded ? "h-auto" : "h-14 lg:h-full"
            )}>
                <div className="bg-card border border-border rounded-xl shadow-sm overflow-hidden flex flex-col h-full">
                    {/* Mobile Toggle Header */}
                    <button
                        onClick={() => setConfigExpanded(!configExpanded)}
                        className="lg:hidden flex items-center justify-between p-4 w-full hover:bg-muted/50 transition-colors"
                    >
                        <div className="flex items-center gap-2 font-semibold">
                            <Sparkles className="w-4 h-4 text-primary" />
                            <span>模型与配置</span>
                        </div>
                        <div className={clsx("transition-transform duration-200", configExpanded ? "rotate-180" : "")}>
                            <ChevronDown className="w-4 h-4" />
                        </div>
                    </button>

                    <div className={clsx(
                        "p-5 space-y-5 overflow-y-auto custom-scrollbar flex-1",
                        !configExpanded && "hidden lg:block"
                    )}>
                        <h3 className="hidden lg:flex font-semibold items-center gap-2 mb-2">
                            <Sparkles className="w-4 h-4 text-primary" />
                            模型选项
                        </h3>

                        <div className="space-y-3">
                            <label className="text-sm font-medium text-muted-foreground">模型</label>
                            <div className="grid grid-cols-2 lg:grid-cols-1 gap-2">
                                {MODELS.map(m => {
                                    const Icon = m.icon
                                    return (
                                        <button
                                            key={m.id}
                                            onClick={() => setModel(m.id)}
                                            className={clsx(
                                                "flex items-center lg:items-center gap-2 lg:gap-3 p-2 lg:p-3 rounded-lg border text-left transition-all",
                                                model === m.id
                                                    ? "border-primary bg-primary/5 ring-1 ring-primary/20"
                                                    : "border-border hover:bg-secondary/50"
                                            )}
                                        >
                                            <div className={clsx("p-1.5 lg:p-2 rounded-md shrink-0", model === m.id ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground")}>
                                                <Icon className="w-3.5 h-3.5 lg:w-4 h-4" />
                                            </div>
                                            <div className="min-w-0">
                                                <div className="font-medium text-xs lg:text-sm truncate">{m.name}</div>
                                                <div className="text-[10px] lg:text-xs text-muted-foreground truncate hidden sm:block lg:block">{m.desc}</div>
                                            </div>
                                        </button>
                                    )
                                })}
                            </div>
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-muted-foreground">账号策略</label>
                            <select
                                className="input-field"
                                value={selectedAccount}
                                onChange={e => setSelectedAccount(e.target.value)}
                            >
                                <option value="">🎲 随机切换 (支持流式预览)</option>
                                {accounts.map((acc, i) => (
                                    <option key={i} value={acc.email || acc.mobile}>
                                        👤 {acc.email || acc.mobile}
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-muted-foreground">API 密钥 (可选)</label>
                            <input
                                type="password"
                                className="input-field font-mono text-xs"
                                placeholder={config.keys?.[0] ? `默认: ${config.keys[0].slice(0, 8)}...` : '输入自定义 API 密钥'}
                                value={apiKey}
                                onChange={e => setApiKey(e.target.value)}
                            />
                        </div>
                    </div>
                </div>
            </div>

            {/* Chat Interface */}
            <div className="lg:col-span-2 flex flex-col bg-card border border-border rounded-xl shadow-sm overflow-hidden min-h-0 flex-1">
                {/* Messages Area */}
                <div className="flex-1 overflow-y-auto p-4 space-y-6 custom-scrollbar">
                    {/* User Message */}
                    <div className="flex gap-4 max-w-3xl mx-auto">
                        <div className="w-8 h-8 rounded-full bg-secondary flex items-center justify-center shrink-0">
                            <User className="w-4 h-4 text-muted-foreground" />
                        </div>
                        <div className="space-y-1 flew-1">
                            <div className="flex items-center gap-2">
                                <span className="font-medium text-sm">用户</span>
                            </div>
                            <div className="bg-secondary/50 rounded-2xl rounded-tl-none px-4 py-3 text-sm border border-border">
                                {message}
                            </div>
                        </div>
                    </div>

                    {/* AI Response */}
                    {(response || isStreaming) && (
                        <div className="flex gap-4 max-w-3xl mx-auto animate-in fade-in slide-in-from-bottom-2">
                            <div className={clsx(
                                "w-8 h-8 rounded-full flex items-center justify-center shrink-0",
                                response?.success !== false ? "bg-primary text-primary-foreground" : "bg-destructive text-destructive-foreground"
                            )}>
                                <Bot className="w-4 h-4" />
                            </div>
                            <div className="space-y-2 flex-1 min-w-0">
                                <div className="flex items-center gap-2">
                                    <span className="font-medium text-sm">DeepSeek</span>
                                    {response && (
                                        <span className={clsx(
                                            "text-[10px] px-1.5 py-0.5 rounded border uppercase font-bold tracking-wider",
                                            response.success ? "border-emerald-500/20 text-emerald-500 bg-emerald-500/5" : "border-destructive/20 text-destructive bg-destructive/5"
                                        )}>
                                            {response.status_code || 'ERROR'}
                                        </span>
                                    )}
                                </div>

                                {(streamingThinking || response?.response?.thinking) && (
                                    <div className="text-xs text-muted-foreground bg-muted/30 border border-border rounded-lg p-3 space-y-1">
                                        <div className="flex items-center gap-1.5 opacity-70 mb-1">
                                            <Sparkles className="w-3 h-3" />
                                            <span>思维链/推理过程</span>
                                        </div>
                                        <div className="whitespace-pre-wrap leading-relaxed opacity-90 font-mono">
                                            {streamingThinking || response?.response?.thinking}
                                        </div>
                                    </div>
                                )}

                                <div className="text-sm leading-relaxed whitespace-pre-wrap">
                                    {!selectedAccount ? (
                                        streamingContent || (response?.error && <span className="text-destructive">{response.error}</span>)
                                    ) : (
                                        response?.response?.message || <span className="text-muted-foreground italic">...</span>
                                    )}
                                    {isStreaming && <span className="inline-block w-1.5 h-4 bg-primary ml-1 align-middle animate-pulse" />}
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Input Area */}
                <div className="p-4 border-t border-border bg-card">
                    <div className="max-w-3xl mx-auto relative">
                        <textarea
                            className="w-full bg-secondary/30 border border-border rounded-xl pl-4 pr-14 py-3 text-sm focus:bg-background focus:ring-1 focus:ring-primary focus:border-primary transition-all resize-none custom-scrollbar"
                            placeholder="输入消息..."
                            rows={3}
                            value={message}
                            onChange={e => setMessage(e.target.value)}
                            onKeyDown={e => {
                                if (e.key === 'Enter' && !e.shiftKey) {
                                    e.preventDefault()
                                    sendTest()
                                }
                            }}
                        />
                        <div className="absolute right-2 bottom-2">
                            {loading && isStreaming ? (
                                <button
                                    onClick={stopGeneration}
                                    className="p-2 bg-destructive text-destructive-foreground rounded-lg hover:opacity-90 transition-opacity"
                                >
                                    <Square className="w-4 h-4 fill-current" />
                                </button>
                            ) : (
                                <button
                                    onClick={sendTest}
                                    disabled={loading || !message.trim()}
                                    className="p-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
                                </button>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}
