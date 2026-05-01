import { s as store_get, u as unsubscribe_stores } from './index2-BkNRUash.js';
import { p as page } from './stores-5e9CBrbK.js';
import './exports-CgQJUv15.js';
import './state.svelte-CF1qhG8z.js';
import './index-CLgSXmZk.js';
import './context-Dj9Hrhuz.js';
import './utils2-BGbnt0UH.js';
import './escaping-CqgfEcN3.js';
import './index-CBcFMcIv.js';

function FormPage($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<div class="form-page-loading svelte-oene30"><div class="form-page-loading-spinner svelte-oene30"></div> <p>Loading form...</p></div>`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    store_get($$store_subs ??= {}, "$page", page).params.formId;
    FormPage($$renderer2);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-D63MkOHy.js.map
