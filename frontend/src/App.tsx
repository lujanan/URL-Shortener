import React, { useState } from "react";

type ShortenResponse = {
  code: string;
  short_url: string;
  long_url: string;
  expire_at?: string | null;
};

type LinkInfoResponse = {
  code: string;
  long_url: string;
  created_at: string;
  expire_at?: string | null;
  click_count: number;
  last_accessed_at?: string | null;
};

const API_BASE_URL =
  (import.meta as any).env?.VITE_API_BASE_URL || "http://localhost:8080";

export const App: React.FC = () => {
  const [url, setUrl] = useState("");
  const [customCode, setCustomCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ShortenResponse | null>(null);

  const [infoCode, setInfoCode] = useState("");
  const [infoLoading, setInfoLoading] = useState(false);
  const [infoError, setInfoError] = useState<string | null>(null);
  const [infoResult, setInfoResult] = useState<LinkInfoResponse | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setResult(null);

    if (!url.trim()) {
      setError("请输入要缩短的 URL");
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/shorten`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          url: url.trim(),
          custom_code: customCode.trim() || undefined
        })
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.message || "创建短链接失败");
      }

      const data: ShortenResponse = await res.json();
      setResult(data);
    } catch (err: any) {
      setError(err.message || "请求失败，请稍后重试");
    } finally {
      setLoading(false);
    }
  };

  const handleFetchInfo = async (e: React.FormEvent) => {
    e.preventDefault();
    setInfoError(null);
    setInfoResult(null);

    if (!infoCode.trim()) {
      setInfoError("请输入短码");
      return;
    }

    setInfoLoading(true);
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/links/${infoCode.trim()}`);
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.message || "查询失败");
      }
      const data: LinkInfoResponse = await res.json();
      setInfoResult(data);
    } catch (err: any) {
      setInfoError(err.message || "请求失败，请稍后重试");
    } finally {
      setInfoLoading(false);
    }
  };

  return (
    <div className="page">
      <div className="card">
        <h1 className="title">短链接生成服务</h1>
        <p className="subtitle">输入长链接，一键生成可分享的短链接。</p>

        <form className="form" onSubmit={handleSubmit}>
          <label className="label">
            长链接 URL
            <input
              className="input"
              type="url"
              placeholder="https://example.com/your-long-url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              required
            />
          </label>

          <label className="label">
            自定义短码（可选）
            <input
              className="input"
              type="text"
              placeholder="例如：promo2026"
              value={customCode}
              onChange={(e) => setCustomCode(e.target.value)}
            />
          </label>

          <button className="button" type="submit" disabled={loading}>
            {loading ? "生成中..." : "生成短链接"}
          </button>
        </form>

        {error && <div className="alert alert-error">{error}</div>}

        {result && (
          <div className="result">
            <h2>生成成功</h2>
            <p>
              短码：<span className="code-chip">{result.code}</span>
            </p>
            <p>
              短链接：
              <a href={result.short_url} target="_blank" rel="noreferrer">
                {result.short_url}
              </a>
            </p>
            <p className="muted">原始链接：{result.long_url}</p>
          </div>
        )}

        <div className="divider" />

        <h2 className="section-title">查询短链信息</h2>
        <p className="subtitle">输入短码，查看原始链接、过期时间、点击次数等。</p>

        <form className="form" onSubmit={handleFetchInfo}>
          <label className="label">
            短码
            <input
              className="input"
              type="text"
              placeholder="例如：abc123"
              value={infoCode}
              onChange={(e) => setInfoCode(e.target.value)}
              required
            />
          </label>

          <button className="button" type="submit" disabled={infoLoading}>
            {infoLoading ? "查询中..." : "查询短链信息"}
          </button>
        </form>

        {infoError && <div className="alert alert-error">{infoError}</div>}

        {infoResult && (
          <div className="result">
            <h2>查询结果</h2>
            <p>
              短链接：
              <a
                href={`${API_BASE_URL}/${infoResult.code}`}
                target="_blank"
                rel="noreferrer"
              >
                {`${API_BASE_URL}/${infoResult.code}`}
              </a>
            </p>
            <p className="muted">原始链接：{infoResult.long_url}</p>
            <p className="muted">
              过期时间：{infoResult.expire_at ? infoResult.expire_at : "永不过期"}
            </p>
            <p className="muted">创建时间：{infoResult.created_at}</p>
            <p className="muted">点击次数：{infoResult.click_count}</p>
            <p className="muted">
              最后访问：{infoResult.last_accessed_at || "尚无访问记录"}
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

