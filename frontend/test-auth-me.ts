import { config } from 'dotenv';
config({ path: '/Users/pertadima/Downloads/moedah-pos/backend/.env' });
async function test() {
  // We need a user to login.
  // First, let's login using standard credentials.
  const res = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: 'admin@moedah.com', password: 'password123' })
  });
  if (!res.ok) {
    console.error('Login failed:', await res.text());
    return;
  }
  const data = await res.json();
  const token = data.data.access_token;
  
  const meRes = await fetch('http://localhost:8080/api/v1/auth/me', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const meData = await meRes.json();
  console.log(JSON.stringify(meData.data.stores, null, 2));
}
test();
