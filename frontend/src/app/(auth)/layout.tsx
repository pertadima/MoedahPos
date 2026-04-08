export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'stretch',
        justifyContent: 'stretch',
        width: '100vw',
        background:
          'radial-gradient(ellipse at 60% 0%, rgba(8,132,246,0.10) 0%, transparent 60%), radial-gradient(ellipse at 10% 100%, rgba(255,167,36,0.06) 0%, transparent 50%), var(--bg-base)',
      }}
    >
      {children}
    </div>
  );
}
