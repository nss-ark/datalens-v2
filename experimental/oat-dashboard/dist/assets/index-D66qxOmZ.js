var O=e=>{throw TypeError(e)};var x=(e,o,t)=>o.has(e)||O("Cannot "+t);var n=(e,o,t)=>(x(e,o,"read from private field"),t?t.call(e):o.get(e)),l=(e,o,t)=>o.has(e)?O("Cannot add the same private member more than once"):o instanceof WeakSet?o.add(e):o.set(e,t),p=(e,o,t,r)=>(x(e,o,"write to private field"),r?r.call(e,t):o.set(e,t),t),v=(e,o,t)=>(x(e,o,"access private method"),t);(function(){const o=document.createElement("link").relList;if(o&&o.supports&&o.supports("modulepreload"))return;for(const s of document.querySelectorAll('link[rel="modulepreload"]'))r(s);new MutationObserver(s=>{for(const i of s)if(i.type==="childList")for(const d of i.addedNodes)d.tagName==="LINK"&&d.rel==="modulepreload"&&r(d)}).observe(document,{childList:!0,subtree:!0});function t(s){const i={};return s.integrity&&(i.integrity=s.integrity),s.referrerPolicy&&(i.referrerPolicy=s.referrerPolicy),s.crossOrigin==="use-credentials"?i.credentials="include":s.crossOrigin==="anonymous"?i.credentials="omit":i.credentials="same-origin",i}function r(s){if(s.ep)return;s.ep=!0;const i=t(s);fetch(s.href,i)}})();var w,E,L;class k extends HTMLElement{constructor(){super(...arguments);l(this,E);l(this,w,!1)}connectedCallback(){n(this,w)||(document.readyState==="loading"?document.addEventListener("DOMContentLoaded",()=>v(this,E,L).call(this),{once:!0}):v(this,E,L).call(this))}init(){}disconnectedCallback(){this.cleanup()}cleanup(){}handleEvent(t){const r=this[`on${t.type}`];r&&r.call(this,t)}emit(t,r=null){return this.dispatchEvent(new CustomEvent(t,{bubbles:!0,composed:!0,cancelable:!0,detail:r}))}getBool(t){return this.hasAttribute(t)}setBool(t,r){r?this.setAttribute(t,""):this.removeAttribute(t)}$(t){return this.querySelector(t)}$$(t){return Array.from(this.querySelectorAll(t))}uid(){return Math.random().toString(36).slice(2,10)}}w=new WeakMap,E=new WeakSet,L=function(){n(this,w)||(p(this,w,!0),this.init())};typeof window<"u"&&(window.OtBase=k),"commandForElement"in HTMLButtonElement.prototype||document.addEventListener("click",e=>{const o=e.target.closest("[commandfor]");if(!o)return;const t=document.getElementById(o.getAttribute("commandfor"));if(!t)return;const r=o.getAttribute("command")||"toggle";t instanceof HTMLDialogElement&&(r==="show-modal"?t.showModal():r==="close"||t.open?t.close():t.showModal())});var a,b,f,A;class C extends k{constructor(){super(...arguments);l(this,f);l(this,a,[]);l(this,b,[])}init(){const t=this.$(':scope > [role="tablist"]');if(p(this,a,t?[...t.querySelectorAll('[role="tab"]')]:[]),p(this,b,this.$$(':scope > [role="tabpanel"]')),n(this,a).length===0||n(this,b).length===0){console.warn("ot-tabs: Missing tab or tabpanel elements");return}n(this,a).forEach((s,i)=>{const d=n(this,b)[i];if(!d)return;const h=s.id||`ot-tab-${this.uid()}`,y=d.id||`ot-panel-${this.uid()}`;s.id=h,d.id=y,s.setAttribute("aria-controls",y),d.setAttribute("aria-labelledby",h),s.addEventListener("click",this),s.addEventListener("keydown",this)});const r=n(this,a).findIndex(s=>s.ariaSelected==="true");v(this,f,A).call(this,r>=0?r:0)}onclick(t){const r=n(this,a).indexOf(t.target.closest('[role="tab"]'));r>=0&&v(this,f,A).call(this,r)}onkeydown(t){const{key:r}=t,s=this.activeIndex;let i=s;switch(r){case"ArrowLeft":t.preventDefault(),i=s-1,i<0&&(i=n(this,a).length-1);break;case"ArrowRight":t.preventDefault(),i=(s+1)%n(this,a).length;break;default:return}v(this,f,A).call(this,i),n(this,a)[i].focus()}get activeIndex(){return n(this,a).findIndex(t=>t.ariaSelected==="true")}set activeIndex(t){t>=0&&t<n(this,a).length&&v(this,f,A).call(this,t)}}a=new WeakMap,b=new WeakMap,f=new WeakSet,A=function(t){n(this,a).forEach((r,s)=>{const i=s===t;r.ariaSelected=String(i),r.tabIndex=i?0:-1}),n(this,b).forEach((r,s)=>{r.hidden=s!==t}),this.emit("ot-tab-change",{index:t,tab:n(this,a)[t]})};customElements.define("ot-tabs",C);var c,u,m;class M extends k{constructor(){super(...arguments);l(this,c);l(this,u);l(this,m)}init(){p(this,c,this.$("menu[popover]")),p(this,u,this.$("[popovertarget]")),!(!n(this,c)||!n(this,u))&&(n(this,c).addEventListener("toggle",this),n(this,c).addEventListener("keydown",this),p(this,m,()=>{const t=n(this,u).getBoundingClientRect();n(this,c).style.top=`${t.bottom}px`,n(this,c).style.left=`${t.left}px`}))}ontoggle(t){var r;t.newState==="open"?(n(this,m).call(this),window.addEventListener("scroll",n(this,m),!0),(r=this.$('[role="menuitem"]'))==null||r.focus(),n(this,u).ariaExpanded="true"):(window.removeEventListener("scroll",n(this,m),!0),n(this,u).ariaExpanded="false",n(this,u).focus())}onkeydown(t){var i,d;if(!t.target.matches('[role="menuitem"]'))return;const r=this.$$('[role="menuitem"]'),s=r.indexOf(t.target);switch(t.key){case"ArrowDown":t.preventDefault(),(i=r[(s+1)%r.length])==null||i.focus();break;case"ArrowUp":t.preventDefault(),(d=r[s-1<0?r.length-1:s-1])==null||d.focus();break}}cleanup(){window.removeEventListener("scroll",n(this,m),!0)}}c=new WeakMap,u=new WeakMap,m=new WeakMap;customElements.define("ot-dropdown",M);const $=window.ot||(window.ot={}),g={},D=4e3,N="top-right";function P(e){if(!g[e]){const o=document.createElement("div");o.className="toast-container",o.setAttribute("popover","manual"),o.setAttribute("data-placement",e),document.body.appendChild(o),g[e]=o}return g[e]}function S(e,o={}){const{placement:t=N,duration:r=D}=o,s=P(t);e.classList.add("toast");let i;return e.onmouseenter=()=>clearTimeout(i),e.onmouseleave=()=>{r>0&&(i=setTimeout(()=>T(e,s),r))},e.setAttribute("data-entering",""),s.appendChild(e),s.showPopover(),requestAnimationFrame(()=>{requestAnimationFrame(()=>{e.removeAttribute("data-entering")})}),r>0&&(i=setTimeout(()=>T(e,s),r)),e}$.toast=function(e,o,t={}){typeof e=="object"&&e!==null&&(t=e,e="");const{variant:r="",...s}=t,i=document.createElement("output");i.className="toast",i.setAttribute("role","status"),r&&i.setAttribute("data-variant",r);const d=o||r[0].toUpperCase()+r.slice(1),h=document.createElement("h6");if(h.className="toast-title",r&&(h.style.color=`var(--${r})`),h.textContent=o||d,i.appendChild(h),e){const y=document.createElement("div");y.className="toast-message",y.textContent=e,i.appendChild(y)}return S(i,s)},$.toastEl=function(e,o={}){let t;return e instanceof HTMLTemplateElement?t=e.content.firstElementChild.cloneNode(!0):typeof e=="string"?t=document.querySelector(e).cloneNode(!0):t=e.cloneNode(!0),t.removeAttribute("id"),S(t,o)};function T(e,o){if(e.hasAttribute("data-exiting"))return;e.setAttribute("data-exiting","");let t=!1;const r=()=>{t||(t=!0,e.remove(),o.children.length||o.hidePopover())};e.addEventListener("transitionend",r,{once:!0}),setTimeout(r,200)}$.toast.clear=function(e){e&&g[e]?(g[e].innerHTML="",g[e].hidePopover()):Object.values(g).forEach(o=>{o.innerHTML="",o.hidePopover()})},document.addEventListener("DOMContentLoaded",()=>{document.querySelectorAll("[title]").forEach(e=>{const o=e.getAttribute("title");o&&(e.setAttribute("data-tooltip",o),e.hasAttribute("aria-label")||e.setAttribute("aria-label",o),e.removeAttribute("title"))})}),document.addEventListener("click",e=>{var t;const o=e.target.closest("[data-sidebar-toggle]");o&&((t=o.closest("[data-sidebar-layout]"))==null||t.toggleAttribute("data-sidebar-open"))});document.querySelector("#app").innerHTML=`
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
