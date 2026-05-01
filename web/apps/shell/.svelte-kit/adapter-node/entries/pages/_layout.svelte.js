import "clsx";
import "../../chunks/ChartTypePicker.svelte_svelte_type_style_lang.js";
import "echarts";
import { s as store_get, h as head, u as unsubscribe_stores, a as attr } from "../../chunks/index2.js";
import { t as themeStore } from "../../chunks/theme.store.js";
function ErpRootLayout($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    let { children } = $$props;
    const theme = store_get($$store_subs ??= {}, "$themeStore", themeStore).mode;
    head("12tpp9o", $$renderer2, ($$renderer3) => {
      $$renderer3.push(`<meta name="color-scheme"${attr("content", theme === "dark" ? "dark" : "light")}/>`);
    });
    children($$renderer2);
    $$renderer2.push(`<!---->`);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}
function _layout($$renderer, $$props) {
  let { children } = $$props;
  ErpRootLayout($$renderer, {
    children: ($$renderer2) => {
      children($$renderer2);
      $$renderer2.push(`<!---->`);
    }
  });
}
export {
  _layout as default
};
