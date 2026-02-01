# -*- coding: utf-8 -*-
"""首页和 WebUI 路由"""
import os
from fastapi import APIRouter, Request
from fastapi.responses import HTMLResponse, FileResponse

from core.config import STATIC_ADMIN_DIR

router = APIRouter()

# 首页 HTML（内嵌避免依赖模板目录）
WELCOME_HTML = """<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DS2API - DeepSeek to OpenAI API</title>
    <meta name="description" content="DS2API - 将 DeepSeek 网页版转换为 OpenAI 兼容 API">
    <meta name="robots" content="noindex, nofollow">
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'%3E%3Cdefs%3E%3ClinearGradient id='g' x1='0%25' y1='0%25' x2='100%25' y2='100%25'%3E%3Cstop offset='0%25' stop-color='%23f59e0b'/%3E%3Cstop offset='100%25' stop-color='%23ef4444'/%3E%3C/linearGradient%3E%3C/defs%3E%3Crect rx='20' width='100' height='100' fill='url(%23g)'/%3E%3Ctext x='50' y='68' font-family='Arial,sans-serif' font-size='48' font-weight='bold' fill='white' text-anchor='middle'%3EDS%3C/text%3E%3C/svg%3E">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1e3a5f 0%, #0f172a 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #fff;
        }
        .container {
            text-align: center;
            padding: 2rem;
        }
        .logo {
            font-size: 4rem;
            font-weight: bold;
            background: linear-gradient(135deg, #f59e0b, #ef4444);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            margin-bottom: 1rem;
        }
        .subtitle {
            color: #94a3b8;
            font-size: 1.2rem;
            margin-bottom: 2rem;
        }
        .links {
            display: flex;
            gap: 1rem;
            justify-content: center;
            flex-wrap: wrap;
        }
        .btn {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            border-radius: 0.5rem;
            text-decoration: none;
            font-weight: 500;
            transition: all 0.2s;
        }
        .btn-primary {
            background: linear-gradient(135deg, #f59e0b, #ef4444);
            color: #fff;
        }
        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 20px rgba(245, 158, 11, 0.4);
        }
        .btn-secondary {
            border: 1px solid #475569;
            color: #94a3b8;
        }
        .btn-secondary:hover {
            border-color: #64748b;
            color: #fff;
        }
        .features {
            margin-top: 3rem;
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1.5rem;
            max-width: 800px;
        }
        .feature {
            background: rgba(255,255,255,0.05);
            padding: 1.5rem;
            border-radius: 0.75rem;
            border: 1px solid rgba(255,255,255,0.1);
        }
        .feature h3 {
            font-size: 1rem;
            margin-bottom: 0.5rem;
        }
        .feature p {
            color: #94a3b8;
            font-size: 0.875rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">DS2API</div>
        <p class="subtitle">DeepSeek to OpenAI Compatible API</p>
        <div class="links">
            <a href="/webui" class="btn btn-primary">🎛️ 管理面板</a>
            <a href="/v1/models" class="btn btn-secondary">📡 API 端点</a>
            <a href="https://github.com/CJackHwang/ds2api" class="btn btn-secondary" target="_blank">📦 GitHub</a>
        </div>
        <div class="features">
            <div class="feature">
                <h3>🚀 OpenAI 兼容</h3>
                <p>完全兼容 OpenAI API 格式</p>
            </div>
            <div class="feature">
                <h3>🔄 多账号轮询</h3>
                <p>Round-Robin 负载均衡</p>
            </div>
            <div class="feature">
                <h3>🧠 深度思考</h3>
                <p>支持 R1 推理模式</p>
            </div>
            <div class="feature">
                <h3>🔍 联网搜索</h3>
                <p>DeepSeek 搜索增强</p>
            </div>
        </div>
    </div>
</body>
</html>"""


@router.get("/")
def index(request: Request):
    return HTMLResponse(content=WELCOME_HTML)


@router.get("/webui")
@router.get("/webui/{path:path}")
async def webui(request: Request, path: str = ""):
    """提供 WebUI 静态文件"""
    # 检查 static/admin 目录是否存在
    if not os.path.isdir(STATIC_ADMIN_DIR):
        return HTMLResponse(
            content="<h1>WebUI not built</h1><p>Run <code>cd webui && npm run build</code> first.</p>",
            status_code=404
        )
    
    # 如果请求的是具体文件（如 js, css）
    if path and "." in path:
        file_path = os.path.join(STATIC_ADMIN_DIR, path)
        if os.path.isfile(file_path):
            return FileResponse(file_path)
        return HTMLResponse(content="Not Found", status_code=404)
    
    # 否则返回 index.html（SPA 路由）
    index_path = os.path.join(STATIC_ADMIN_DIR, "index.html")
    if os.path.isfile(index_path):
        return FileResponse(index_path)
    
    return HTMLResponse(content="index.html not found", status_code=404)
