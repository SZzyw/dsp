import { useState, useEffect } from 'react'
import {
    LayoutDashboard,
    Key,
    Upload,
    Cloud,
    LogOut,
    Menu,
    X,
    Server,
    Users
} from 'lucide-react'
import clsx from 'clsx'

import AccountManager from './components/AccountManager'
import ApiTester from './components/ApiTester'
import BatchImport from './components/BatchImport'
import VercelSync from './components/VercelSync'
import Login from './components/Login'

const NAV_ITEMS = [
    { id: 'accounts', label: '账号管理', icon: Users, description: '管理 DeepSeek 账号池' },
    { id: 'test', label: 'API 测试', icon: Server, description: '测试 API 连接与响应' },
    { id: 'import', label: '批量导入', icon: Upload, description: '批量导入账号配置' },
    { id: 'vercel', label: 'Vercel 同步', icon: Cloud, description: '同步配置到 Vercel' },
]

export default function App() {
    const [activeTab, setActiveTab] = useState('accounts')
    const [config, setConfig] = useState({ keys: [], accounts: [] })
    const [loading, setLoading] = useState(true)
    const [message, setMessage] = useState(null)
    const [token, setToken] = useState(null)
    const [authChecking, setAuthChecking] = useState(true)
    const [sidebarOpen, setSidebarOpen] = useState(false)

    // 检查已存储的 Token
    useEffect(() => {
        const checkAuth = async () => {
            const storedToken = localStorage.getItem('ds2api_token') || sessionStorage.getItem('ds2api_token')
            const expiresAt = parseInt(localStorage.getItem('ds2api_token_expires') || sessionStorage.getItem('ds2api_token_expires') || '0')

            if (storedToken && expiresAt > Date.now()) {
                try {
                    const res = await fetch('/admin/verify', {
                        headers: { 'Authorization': `Bearer ${storedToken}` }
                    })
                    if (res.ok) {
                        setToken(storedToken)
                    } else {
                        handleLogout()
                    }
                } catch {
                    setToken(storedToken)
                }
            }
            setAuthChecking(false)
        }
        checkAuth()
    }, [])

    const authFetch = async (url, options = {}) => {
        const headers = {
            ...options.headers,
            'Authorization': `Bearer ${token}`
        }
        const res = await fetch(url, { ...options, headers })

        if (res.status === 401) {
            handleLogout()
            throw new Error('认证已过期，请重新登录')
        }
        return res
    }

    const fetchConfig = async () => {
        if (!token) return
        try {
            setLoading(true)
            const res = await authFetch('/admin/config')
            if (res.ok) {
                const data = await res.json()
                setConfig(data)
            }
        } catch (e) {
            console.error('获取配置失败:', e)
            showMessage('error', e.message)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        if (token) {
            fetchConfig()
        }
    }, [token])

    const showMessage = (type, text) => {
        setMessage({ type, text })
        setTimeout(() => setMessage(null), 5000)
    }

    const handleLogin = (newToken) => {
        setToken(newToken)
    }

    const handleLogout = () => {
        setToken(null)
        localStorage.removeItem('ds2api_token')
        localStorage.removeItem('ds2api_token_expires')
        sessionStorage.removeItem('ds2api_token')
        sessionStorage.removeItem('ds2api_token_expires')
    }

    const renderTab = () => {
        switch (activeTab) {
            case 'accounts':
                return <AccountManager config={config} onRefresh={fetchConfig} onMessage={showMessage} authFetch={authFetch} />
            case 'test':
                return <ApiTester config={config} onMessage={showMessage} authFetch={authFetch} />
            case 'import':
                return <BatchImport onRefresh={fetchConfig} onMessage={showMessage} authFetch={authFetch} />
            case 'vercel':
                return <VercelSync onMessage={showMessage} authFetch={authFetch} />
            default:
                return null
        }
    }

    if (authChecking) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-background">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
                    <p className="text-muted-foreground animate-pulse">检查登录状态...</p>
                </div>
            </div>
        )
    }

    if (!token) {
        return (
            <div className="min-h-screen flex flex-col bg-background relative overflow-hidden">
                {/* Background decorative elements */}
                <div className="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none z-0">
                    <div className="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-primary/5 rounded-full blur-[120px]"></div>
                    <div className="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-accent/5 rounded-full blur-[120px]"></div>
                </div>

                {message && (
                    <div className={clsx(
                        "fixed top-4 right-4 z-50 px-4 py-3 rounded-lg shadow-lg border animate-in slide-in-from-top-2 fade-in",
                        message.type === 'error' ? "bg-destructive/10 border-destructive/20 text-destructive" :
                            "bg-primary/10 border-primary/20 text-primary"
                    )}>
                        {message.text}
                    </div>
                )}
                <Login onLogin={handleLogin} onMessage={showMessage} />
            </div>
        )
    }

    return (
        <div className="flex h-screen bg-background overflow-hidden text-foreground">
            {/* Mobile Sidebar Overlay */}
            {sidebarOpen && (
                <div
                    className="fixed inset-0 bg-background/80 backdrop-blur-sm z-40 lg:hidden"
                    onClick={() => setSidebarOpen(false)}
                />
            )}

            {/* Sidebar */}
            <aside className={clsx(
                "fixed lg:static inset-y-0 left-0 z-50 w-64 bg-card border-r border-border transition-transform duration-200 ease-in-out lg:transform-none flex flex-col",
                sidebarOpen ? "translate-x-0" : "-translate-x-full"
            )}>
                <div className="p-6 border-b border-border">
                    <div className="flex items-center gap-2 font-bold text-xl text-primary">
                        <LayoutDashboard className="w-6 h-6" />
                        <span>DS2API 管理面板</span>
                    </div>
                    <p className="text-xs text-muted-foreground mt-1 ml-8">V1.0.0 控制面板</p>
                </div>

                <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
                    {NAV_ITEMS.map((item) => {
                        const Icon = item.icon
                        const isActive = activeTab === item.id
                        return (
                            <button
                                key={item.id}
                                onClick={() => {
                                    setActiveTab(item.id)
                                    setSidebarOpen(false)
                                }}
                                className={clsx(
                                    "w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 group",
                                    isActive
                                        ? "bg-primary text-primary-foreground shadow-md shadow-primary/20"
                                        : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                                )}
                            >
                                <Icon className={clsx("w-4 h-4", isActive ? "text-primary-foreground" : "text-muted-foreground group-hover:text-accent-foreground")} />
                                {item.label}
                            </button>
                        )
                    })}
                </nav>

                <div className="p-4 border-t border-border bg-card/50">
                    <div className="space-y-4">
                        <div className="flex items-center justify-between text-sm">
                            <span className="text-muted-foreground">API 状态</span>
                            <span className="flex items-center gap-1.5 text-xs font-medium text-green-500 bg-green-500/10 px-2 py-0.5 rounded-full">
                                <span className="w-1.5 h-1.5 rounded-full bg-green-500 animate-pulse"></span>
                                在线
                            </span>
                        </div>
                        <div className="grid grid-cols-2 gap-2">
                            <div className="bg-background rounded p-2 border border-border">
                                <div className="text-xs text-muted-foreground">账号</div>
                                <div className="text-lg font-bold">{config.accounts?.length || 0}</div>
                            </div>
                            <div className="bg-background rounded p-2 border border-border">
                                <div className="text-xs text-muted-foreground">API 密钥</div>
                                <div className="text-lg font-bold">{config.keys?.length || 0}</div>
                            </div>
                        </div>
                        <button
                            onClick={handleLogout}
                            className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-md border border-border text-sm font-medium text-muted-foreground hover:bg-destructive/10 hover:text-destructive hover:border-destructive/20 transition-colors"
                        >
                            <LogOut className="w-4 h-4" />
                            退出登录
                        </button>
                    </div>
                </div>
            </aside>

            {/* Main Content */}
            <main className="flex-1 flex flex-col min-w-0 overflow-hidden relative">
                {/* Mobile Header */}
                <header className="lg:hidden flex items-center justify-between p-4 border-b border-border bg-card">
                    <span className="font-semibold">{NAV_ITEMS.find(n => n.id === activeTab)?.label}</span>
                    <button
                        onClick={() => setSidebarOpen(true)}
                        className="p-2 -mr-2 text-muted-foreground hover:text-foreground"
                    >
                        <Menu className="w-5 h-5" />
                    </button>
                </header>

                {/* Content Area */}
                <div className="flex-1 overflow-auto bg-background/50 p-2 sm:p-4 lg:p-8">
                    <div className="max-w-6xl mx-auto space-y-4 lg:space-y-6">
                        <div className="hidden lg:block mb-8">
                            <h1 className="text-3xl font-bold tracking-tight mb-2">
                                {NAV_ITEMS.find(n => n.id === activeTab)?.label}
                            </h1>
                            <p className="text-muted-foreground">
                                {NAV_ITEMS.find(n => n.id === activeTab)?.description}
                            </p>
                        </div>

                        {message && (
                            <div className={clsx(
                                "p-4 rounded-lg border flex items-center gap-3 animate-in fade-in slide-in-from-top-2",
                                message.type === 'error' ? "bg-destructive/10 border-destructive/20 text-destructive" :
                                    "bg-emerald-500/10 border-emerald-500/20 text-emerald-500"
                            )}>
                                {message.type === 'error' ? <X className="w-5 h-5" /> : <div className="w-5 h-5 rounded-full border-2 border-emerald-500 flex items-center justify-center text-[10px]">✓</div>}
                                {message.text}
                            </div>
                        )}

                        <div className="animate-in fade-in duration-500">
                            {loading ? (
                                <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
                                    <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin mb-4"></div>
                                    <p>正在加载数据，请稍候...</p>
                                </div>
                            ) : (
                                renderTab()
                            )}
                        </div>
                    </div>
                </div>
            </main>
        </div>
    )
}
