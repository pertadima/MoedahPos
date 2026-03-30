export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div style={{
      minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center',
      background: 'radial-gradient(ellipse at 60% 0%, rgba(16,185,129,0.08) 0%, transparent 60%), var(--bg-base)',
      padding: 16,
    }}>
      {children}
    </div>
  );
}
