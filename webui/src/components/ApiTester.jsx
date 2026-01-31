import { useState } from 'react'

const MODELS = [
    { id: 'deepseek-chat', name: 'deepseek-chat' },
    { id: 'deepseek-reasoner', name: 'deepseek-reasoner' },
    { id: 'deepseek-chat-search', name: 'deepseek-chat-search' },
    { id: 'deepseek-reasoner-search', name: 'deepseek-reasoner-search' },
]

export default function ApiTester({ config, onMessage }) {
    const [model, setModel] = useState('deepseek-chat')
    const [message, setMessage] = useState('你好，请用一句话介绍你自己。')
    const [apiKey, setApiKey] = useState('')
    const [response, setResponse] = useState(null)
    const [loading, setLoading] = useState(false)

    const testApi = async () => {
        setLoading(true)
        setResponse(null)
        try {
            const res = await fetch('/admin/test', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model,
                    message,
                    api_key: apiKey || (config.keys?.[0] || ''),
                }),
            })
            const data = await res.json()
            setResponse(data)
            if (data.success) {
                onMessage('success', 'API 调用成功')
            } else {
                onMessage('error', data.error || 'API 调用失败')
            }
        } catch (e) {
            onMessage('error', '网络错误')
            setResponse({ error: e.message })
        } finally {
            setLoading(false)
        }
    }

    const directTest = async () => {
        setLoading(true)
        setResponse(null)
        try {
            const key = apiKey || (config.keys?.[0] || '')
            if (!key) {
                onMessage('error', '请提供 API Key')
                setLoading(false)
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
                    stream: false,
                }),
            })
            const data = await res.json()
            setResponse({
                success: res.ok,
                status_code: res.status,
                response: data,
            })
            if (res.ok) {
                onMessage('success', 'API 调用成功')
            } else {
                onMessage('error', data.error || 'API 调用失败')
            }
        } catch (e) {
            onMessage('error', '网络错误')
            setResponse({ error: e.message })
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="section">
            <div className="card">
                <div className="card-title" style={{ marginBottom: '1rem' }}>🧪 API 测试</div>

                <div className="form-group">
                    <label className="form-label">模型</label>
                    <select
                        className="form-input"
                        value={model}
                        onChange={e => setModel(e.target.value)}
                    >
                        {MODELS.map(m => (
                            <option key={m.id} value={m.id}>{m.name}</option>
                        ))}
                    </select>
                </div>

                <div className="form-group">
                    <label className="form-label">API Key（留空使用第一个配置的 Key）</label>
                    <input
                        type="text"
                        className="form-input"
                        placeholder={config.keys?.[0] ? `默认: ${config.keys[0].slice(0, 8)}...` : '请先添加 API Key'}
                        value={apiKey}
                        onChange={e => setApiKey(e.target.value)}
                    />
                </div>

                <div className="form-group">
                    <label className="form-label">消息内容</label>
                    <textarea
                        className="form-input"
                        value={message}
                        onChange={e => setMessage(e.target.value)}
                        placeholder="输入测试消息..."
                    />
                </div>

                <div className="btn-group">
                    <button className="btn btn-primary" onClick={directTest} disabled={loading}>
                        {loading ? <span className="loading"></span> : '🚀 发送请求'}
                    </button>
                </div>
            </div>

            {response && (
                <div className="card">
                    <div className="card-header">
                        <span className="card-title">响应结果</span>
                        <span className={`badge ${response.success ? 'badge-success' : 'badge-error'}`}>
                            {response.success ? '成功' : '失败'} {response.status_code && `(${response.status_code})`}
                        </span>
                    </div>
                    <div className="code-block">
                        {JSON.stringify(response.response || response.error, null, 2)}
                    </div>

                    {response.success && response.response?.choices?.[0]?.message?.content && (
                        <div style={{ marginTop: '1rem' }}>
                            <div className="form-label">AI 回复：</div>
                            <div style={{
                                padding: '1rem',
                                background: 'var(--bg-tertiary)',
                                borderRadius: 'var(--radius)',
                                whiteSpace: 'pre-wrap'
                            }}>
                                {response.response.choices[0].message.content}
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    )
}
