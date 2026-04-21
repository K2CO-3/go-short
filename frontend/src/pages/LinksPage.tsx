import { FormEvent, useEffect, useState } from "react";
import { api } from "../lib/api";
import { getToken } from "../lib/auth";
import type { LinkItem } from "../lib/types";

export function LinksPage() {
  const token = getToken()!;
  const [items, setItems] = useState<LinkItem[]>([]);
  const [url, setUrl] = useState("");
  const [shortCode, setShortCode] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  async function load() {
    try {
      const res = await api.getLinks(token, 1);
      setItems(res.links);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function create(e: FormEvent) {
    e.preventDefault();
    setError("");
    setMessage("");
    try {
      const res = await api.createLink(token, {
        url,
        short_code: shortCode || undefined
      });
      setMessage(res.message || "创建成功");
      setUrl("");
      setShortCode("");
      await load();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function remove(id: number) {
    try {
      await api.deleteLink(token, id);
      await load();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  function displayShortUrl(code: string): string {
    return `${window.location.origin}/code/${code}`;
  }

  return (
    <section>
      <h2>短链管理</h2>
      <form className="card" onSubmit={create}>
        <input value={url} onChange={(e) => setUrl(e.target.value)} placeholder="原始URL" />
        <input
          value={shortCode}
          onChange={(e) => setShortCode(e.target.value)}
          placeholder="自定义短码(可选)"
        />
        <button type="submit">创建</button>
        {message && <p className="ok">{message}</p>}
        {error && <p className="error">{error}</p>}
      </form>

      <div className="card">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>短码</th>
              <th>短链</th>
              <th>原始地址</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {items.map((i) => (
              <tr key={i.link_id}>
                <td>{i.link_id}</td>
                <td>{i.short_code}</td>
                <td>
                  <a href={displayShortUrl(i.short_code)} target="_blank" rel="noreferrer">
                    {displayShortUrl(i.short_code)}
                  </a>
                </td>
                <td>{i.original_url}</td>
                <td>
                  <button onClick={() => void remove(i.link_id)}>删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
