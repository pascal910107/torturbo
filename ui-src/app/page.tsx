'use client'
/* eslint-disable @typescript-eslint/no-explicit-any */

import useSWR from 'swr'

const fetcher = (url: string) => fetch(url).then(res => res.json())

export default function Home() {
  const { data } = useSWR('/api/status', fetcher, { refreshInterval: 5000 })

  return (
    <main className="min-h-screen bg-gray-950 text-gray-100 p-6">
      <h1 className="text-2xl font-bold mb-6">TorTurbo Dashboard</h1>
      <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        {data?.circuits?.map((c: any, idx: number) => (
          <div key={idx} className="rounded-2xl bg-gray-800 p-4 shadow-md">
            <h2 className="text-lg font-semibold mb-2">Circuit #{idx+1}</h2>
            <p>RTT: <span className="font-mono">{c.rtt}</span></p>
          </div>
        )) ?? <p>Loading…</p>}
      </div>
    </main>
  );
}
