import '@knadh/oat/oat.min.css';
import '@knadh/oat/oat.min.js';
import './styles/main.css';

document.querySelector('#app').innerHTML = `
  <aside class="sidebar">
    <div style="font-weight: bold; font-size: 1.5rem; color: var(--primary); margin-bottom: 2rem;">
      OatDash
    </div>
    <nav>
      <a href="#" style="display: block; padding: 0.75rem; color: var(--foreground); text-decoration: none; border-radius: 4px; background: var(--secondary); margin-bottom: 0.5rem;">Dashboard</a>
      <a href="#" style="display: block; padding: 0.75rem; color: var(--muted-foreground); text-decoration: none;">Analytics</a>
      <a href="#" style="display: block; padding: 0.75rem; color: var(--muted-foreground); text-decoration: none;">Settings</a>
    </nav>
  </aside>

  <main class="main-content">
    <header class="header">
      <div style="font-weight: bold;">Overview</div>
      <div style="display: flex; gap: 1rem; align-items: center;">
        <input type="text" placeholder="Search..." style="background: var(--input); border: none; padding: 0.5rem; border-radius: 4px; color: var(--foreground);">
        <div style="width: 32px; height: 32px; background: var(--primary); border-radius: 50%;"></div>
      </div>
    </header>

    <div class="scroll-area">
      <div class="dashboard-grid">
        <div class="stat-card">
          <div class="stat-label">Total Revenue</div>
          <div class="stat-value">$45,231.89</div>
          <div style="color: var(--success); font-size: 0.875rem; margin-top: 0.5rem;">+20.1% from last month</div>
        </div>
        <div class="stat-card">
           <div class="stat-label">Active Users</div>
          <div class="stat-value">+2350</div>
          <div style="color: var(--success); font-size: 0.875rem; margin-top: 0.5rem;">+180.1% from last month</div>
        </div>
        <div class="stat-card">
           <div class="stat-label">Sales</div>
          <div class="stat-value">+12,234</div>
           <div style="color: var(--success); font-size: 0.875rem; margin-top: 0.5rem;">+19% from last month</div>
        </div>
         <div class="stat-card">
           <div class="stat-label">Active Now</div>
          <div class="stat-value">+573</div>
           <div style="color: var(--success); font-size: 0.875rem; margin-top: 0.5rem;">+201 since last hour</div>
        </div>
      </div>

      <div style="background: var(--card); border: 1px solid var(--border); border-radius: 8px; padding: 1.5rem;">
        <h3 style="margin-top: 0; margin-bottom: 1.5rem;">Recent Sales</h3>
        <table style="width: 100%; border-collapse: collapse; text-align: left;">
          <thead>
            <tr style="border-bottom: 1px solid var(--border);">
              <th style="padding: 0.75rem; color: var(--muted-foreground); font-weight: normal;">Customer</th>
              <th style="padding: 0.75rem; color: var(--muted-foreground); font-weight: normal;">Status</th>
              <th style="padding: 0.75rem; color: var(--muted-foreground); font-weight: normal;">Amount</th>
            </tr>
          </thead>
          <tbody>
            <tr style="border-bottom: 1px solid var(--border);">
              <td style="padding: 0.75rem;">Olivia Martin</td>
              <td style="padding: 0.75rem;"><span style="color: var(--success);">Keep</span></td>
              <td style="padding: 0.75rem;">$1,999.00</td>
            </tr>
             <tr style="border-bottom: 1px solid var(--border);">
              <td style="padding: 0.75rem;">Jackson Lee</td>
              <td style="padding: 0.75rem;"><span style="color: var(--warning);">Pending</span></td>
              <td style="padding: 0.75rem;">$39.00</td>
            </tr>
             <tr style="border-bottom: 1px solid var(--border);">
              <td style="padding: 0.75rem;">Isabella Nguyen</td>
              <td style="padding: 0.75rem;"><span style="color: var(--danger);">Failed</span></td>
              <td style="padding: 0.75rem;">$299.00</td>
            </tr>
             <tr>
              <td style="padding: 0.75rem;">William Kim</td>
              <td style="padding: 0.75rem;"><span style="color: var(--success);">Paid</span></td>
              <td style="padding: 0.75rem;">$99.00</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </main>
`;
