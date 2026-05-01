
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	export interface AppTypes {
		RouteId(): "/(auth)" | "/(app)" | "/" | "/%28app%29" | "/%28app%29/land" | "/%28app%29/land/%5Bcategory%5D" | "/(app)/approvals" | "/(app)/dashboard" | "/(app)/forms" | "/(app)/forms/[module]" | "/(app)/forms/[module]/[formId]" | "/(app)/forms/[module]/[formId]/submissions" | "/(auth)/login" | "/(app)/[domain]" | "/(app)/[domain]/[entity]";
		RouteParams(): {
			"/(app)/forms/[module]": { module: string };
			"/(app)/forms/[module]/[formId]": { module: string; formId: string };
			"/(app)/forms/[module]/[formId]/submissions": { module: string; formId: string };
			"/(app)/[domain]": { domain: string };
			"/(app)/[domain]/[entity]": { domain: string; entity: string }
		};
		LayoutParams(): {
			"/(auth)": Record<string, never>;
			"/(app)": { module?: string; formId?: string; domain?: string; entity?: string };
			"/": { module?: string; formId?: string; domain?: string; entity?: string };
			"/%28app%29": Record<string, never>;
			"/%28app%29/land": Record<string, never>;
			"/%28app%29/land/%5Bcategory%5D": Record<string, never>;
			"/(app)/approvals": Record<string, never>;
			"/(app)/dashboard": Record<string, never>;
			"/(app)/forms": { module?: string; formId?: string };
			"/(app)/forms/[module]": { module: string; formId?: string };
			"/(app)/forms/[module]/[formId]": { module: string; formId: string };
			"/(app)/forms/[module]/[formId]/submissions": { module: string; formId: string };
			"/(auth)/login": Record<string, never>;
			"/(app)/[domain]": { domain: string; entity?: string };
			"/(app)/[domain]/[entity]": { domain: string; entity: string }
		};
		Pathname(): "/" | "/%28app%29" | "/%28app%29/" | "/%28app%29/land" | "/%28app%29/land/" | "/%28app%29/land/%5Bcategory%5D" | "/%28app%29/land/%5Bcategory%5D/" | "/approvals" | "/approvals/" | "/dashboard" | "/dashboard/" | "/forms" | "/forms/" | `/forms/${string}` & {} | `/forms/${string}/` & {} | `/forms/${string}/${string}` & {} | `/forms/${string}/${string}/` & {} | `/forms/${string}/${string}/submissions` & {} | `/forms/${string}/${string}/submissions/` & {} | "/login" | "/login/" | `/${string}` & {} | `/${string}/` & {} | `/${string}/${string}` & {} | `/${string}/${string}/` & {};
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): "/favicon.png" | "/icons.svg" | "/manifest.json" | "/offline.html" | "/sw.js" | string & {};
	}
}