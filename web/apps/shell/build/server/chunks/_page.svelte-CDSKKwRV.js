import { h as head, s as store_get, e as ensure_array_like, b as attr_class, u as unsubscribe_stores } from './index2-BkNRUash.js';
import { e as escape_html } from './escaping-CqgfEcN3.js';
import { a as authStore } from './auth.store-D0sp4P0v.js';
import './context-Dj9Hrhuz.js';
import './utils2-BGbnt0UH.js';
import './index-CBcFMcIv.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    const metrics = [
      {
        label: "Total Revenue",
        value: "$125,430",
        change: "+12.5%",
        changeType: "positive"
      },
      {
        label: "Orders",
        value: "1,234",
        change: "+8.2%",
        changeType: "positive"
      },
      {
        label: "Customers",
        value: "856",
        change: "+5.1%",
        changeType: "positive"
      },
      {
        label: "Pending Invoices",
        value: "23",
        change: "-3.4%",
        changeType: "negative"
      }
    ];
    const recentActivities = [
      {
        id: 1,
        action: "New order received",
        time: "5 minutes ago",
        icon: "order"
      },
      {
        id: 2,
        action: "Payment confirmed for INV-2024-001",
        time: "1 hour ago",
        icon: "payment"
      },
      {
        id: 3,
        action: "New customer registered",
        time: "2 hours ago",
        icon: "customer"
      },
      {
        id: 4,
        action: "Stock alert: Widget A running low",
        time: "3 hours ago",
        icon: "alert"
      },
      {
        id: 5,
        action: "Invoice INV-2024-002 sent",
        time: "4 hours ago",
        icon: "invoice"
      }
    ];
    head("1tyszyy", $$renderer2, ($$renderer3) => {
      $$renderer3.title(($$renderer4) => {
        $$renderer4.push(`<title>Dashboard - samavāya ERP</title>`);
      });
    });
    $$renderer2.push(`<div class="dashboard svelte-1tyszyy"><section class="welcome-section svelte-1tyszyy"><h2 class="welcome-title svelte-1tyszyy">Welcome back, ${escape_html(store_get($$store_subs ??= {}, "$authStore", authStore).user?.displayName || "User")}!</h2> <p class="welcome-subtitle svelte-1tyszyy">Here's what's happening with your business today.</p></section> <section class="metrics-section"><div class="metrics-grid svelte-1tyszyy"><!--[-->`);
    const each_array = ensure_array_like(metrics);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let metric = each_array[$$index];
      $$renderer2.push(`<div class="metric-card svelte-1tyszyy"><div class="metric-header svelte-1tyszyy"><span class="metric-label svelte-1tyszyy">${escape_html(metric.label)}</span> <span${attr_class("metric-change svelte-1tyszyy", void 0, {
        "positive": metric.changeType === "positive",
        "negative": metric.changeType === "negative"
      })}>${escape_html(metric.change)}</span></div> <div class="metric-value svelte-1tyszyy">${escape_html(metric.value)}</div></div>`);
    }
    $$renderer2.push(`<!--]--></div></section> <div class="content-grid svelte-1tyszyy"><section class="activity-section card svelte-1tyszyy"><h3 class="section-title svelte-1tyszyy">Recent Activity</h3> <ul class="activity-list svelte-1tyszyy"><!--[-->`);
    const each_array_1 = ensure_array_like(recentActivities);
    for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
      let activity = each_array_1[$$index_1];
      $$renderer2.push(`<li class="activity-item svelte-1tyszyy"><div class="activity-icon svelte-1tyszyy"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="svelte-1tyszyy">`);
      if (activity.icon === "order") {
        $$renderer2.push("<!--[-->");
        $$renderer2.push(`<path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z"></path><line x1="3" y1="6" x2="21" y2="6"></line><path d="M16 10a4 4 0 01-8 0"></path>`);
      } else {
        $$renderer2.push("<!--[!-->");
        if (activity.icon === "payment") {
          $$renderer2.push("<!--[-->");
          $$renderer2.push(`<rect x="1" y="4" width="22" height="16" rx="2" ry="2"></rect><line x1="1" y1="10" x2="23" y2="10"></line>`);
        } else {
          $$renderer2.push("<!--[!-->");
          if (activity.icon === "customer") {
            $$renderer2.push("<!--[-->");
            $$renderer2.push(`<path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"></path><circle cx="12" cy="7" r="4"></circle>`);
          } else {
            $$renderer2.push("<!--[!-->");
            if (activity.icon === "alert") {
              $$renderer2.push("<!--[-->");
              $$renderer2.push(`<path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line>`);
            } else {
              $$renderer2.push("<!--[!-->");
              $$renderer2.push(`<path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline>`);
            }
            $$renderer2.push(`<!--]-->`);
          }
          $$renderer2.push(`<!--]-->`);
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]--></svg></div> <div class="activity-content svelte-1tyszyy"><p class="activity-action svelte-1tyszyy">${escape_html(activity.action)}</p> <span class="activity-time svelte-1tyszyy">${escape_html(activity.time)}</span></div></li>`);
    }
    $$renderer2.push(`<!--]--></ul> <a href="/activity" class="view-all-link svelte-1tyszyy">View all activity</a></section> <section class="quick-actions-section card svelte-1tyszyy"><h3 class="section-title svelte-1tyszyy">Quick Actions</h3> <div class="quick-actions-grid svelte-1tyszyy"><a href="/sales/orders/new" class="quick-action svelte-1tyszyy"><svg class="quick-action-icon svelte-1tyszyy" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg> <span class="svelte-1tyszyy">New Order</span></a> <a href="/finance/invoices/new" class="quick-action svelte-1tyszyy"><svg class="quick-action-icon svelte-1tyszyy" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="12" y1="18" x2="12" y2="12"></line><line x1="9" y1="15" x2="15" y2="15"></line></svg> <span class="svelte-1tyszyy">New Invoice</span></a> <a href="/masters/customers/new" class="quick-action svelte-1tyszyy"><svg class="quick-action-icon svelte-1tyszyy" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"></path><circle cx="8.5" cy="7" r="4"></circle><line x1="20" y1="8" x2="20" y2="14"></line><line x1="23" y1="11" x2="17" y2="11"></line></svg> <span class="svelte-1tyszyy">Add Customer</span></a> <a href="/inventory/items/new" class="quick-action svelte-1tyszyy"><svg class="quick-action-icon svelte-1tyszyy" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"></path><line x1="12" y1="8" x2="12" y2="16"></line><line x1="8" y1="12" x2="16" y2="12"></line></svg> <span class="svelte-1tyszyy">Add Item</span></a></div></section></div></div>`);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-CDSKKwRV.js.map
