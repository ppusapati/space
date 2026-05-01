
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * Environment variables [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env`. Like [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), this module cannot be imported into client-side code. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * _Unlike_ [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), the values exported from this module are statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * ```ts
 * import { API_KEY } from '$env/static/private';
 * ```
 * 
 * Note that all environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * 
 * ```
 * MY_FEATURE_FLAG=""
 * ```
 * 
 * You can override `.env` values from the command line like so:
 * 
 * ```sh
 * MY_FEATURE_FLAG="enabled" npm run dev
 * ```
 */
declare module '$env/static/private' {
	export const VITE_API_URL: string;
	export const ACSetupSvcPort: string;
	export const ACSvcPort: string;
	export const AI_AGENT: string;
	export const ALLUSERSPROFILE: string;
	export const APPDATA: string;
	export const APPLICATION_INSIGHTS_NO_STATSBEAT: string;
	export const ChocolateyInstall: string;
	export const ChocolateyLastPathUpdate: string;
	export const ChocolateyToolsLocation: string;
	export const CHROME_CRASHPAD_PIPE_NAME: string;
	export const CLAUDECODE: string;
	export const CLAUDE_AGENT_SDK_VERSION: string;
	export const CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING: string;
	export const CLAUDE_CODE_ENTRYPOINT: string;
	export const CLAUDE_CODE_EXECPATH: string;
	export const CLAUDE_CODE_MAX_OUTPUT_TOKENS: string;
	export const COMMONPROGRAMFILES: string;
	export const CommonProgramW6432: string;
	export const COMPUTERNAME: string;
	export const COMSPEC: string;
	export const COPILOT_OTEL_ENABLED: string;
	export const COPILOT_OTEL_EXPORTER_TYPE: string;
	export const COPILOT_OTEL_FILE_EXPORTER_PATH: string;
	export const COREPACK_ENABLE_AUTO_PIN: string;
	export const DriverData: string;
	export const EFC_10092_1262719628: string;
	export const EFC_10092_1592913036: string;
	export const EFC_10092_2283032206: string;
	export const EFC_10092_3789132940: string;
	export const ELECTRON_RUN_AS_NODE: string;
	export const EnableLog: string;
	export const EXEPATH: string;
	export const GIT_EDITOR: string;
	export const GIT_LFS_PATH: string;
	export const GOPATH: string;
	export const GOROOT: string;
	export const HOME: string;
	export const HOMEDRIVE: string;
	export const HOMEPATH: string;
	export const IGCCSVC_DB: string;
	export const INIT_CWD: string;
	export const JAVA_HOME: string;
	export const LM_STUDIO_API_BASE: string;
	export const LM_STUDIO_API_KEY: string;
	export const LOCALAPPDATA: string;
	export const LOGONSERVER: string;
	export const MCP_CONNECTION_NONBLOCKING: string;
	export const MSYSTEM: string;
	export const NODE: string;
	export const NoDefaultCurrentDirectoryInExePath: string;
	export const NODE_ENV: string;
	export const npm_command: string;
	export const npm_config_auto_install_peers: string;
	export const npm_config_cache: string;
	export const npm_config_frozen_lockfile: string;
	export const npm_config_manage_package_manager_versions: string;
	export const npm_config_msvs_version: string;
	export const npm_config_node_gyp: string;
	export const npm_config_node_linker: string;
	export const npm_config_python: string;
	export const npm_config_registry: string;
	export const npm_config_user_agent: string;
	export const npm_config_verify_deps_before_run: string;
	export const npm_config__jsr_registry: string;
	export const npm_execpath: string;
	export const npm_lifecycle_event: string;
	export const npm_lifecycle_script: string;
	export const npm_node_execpath: string;
	export const npm_package_json: string;
	export const npm_package_name: string;
	export const npm_package_version: string;
	export const NUMBER_OF_PROCESSORS: string;
	export const OLDPWD: string;
	export const OneDrive: string;
	export const OS: string;
	export const OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE: string;
	export const OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT: string;
	export const PATH: string;
	export const PATHEXT: string;
	export const PLINK_PROTOCOL: string;
	export const pnpm_config_verify_deps_before_run: string;
	export const PNPM_HOME: string;
	export const PNPM_SCRIPT_SRC_DIR: string;
	export const POWERSHELL_DISTRIBUTION_CHANNEL: string;
	export const PROCESSOR_ARCHITECTURE: string;
	export const PROCESSOR_IDENTIFIER: string;
	export const PROCESSOR_LEVEL: string;
	export const PROCESSOR_REVISION: string;
	export const ProgramData: string;
	export const PROGRAMFILES: string;
	export const ProgramW6432: string;
	export const PROMPT: string;
	export const PSModulePath: string;
	export const PUBLIC: string;
	export const PWD: string;
	export const RlsSvcPort: string;
	export const SESSIONNAME: string;
	export const SHELL: string;
	export const SHLVL: string;
	export const SYSTEMDRIVE: string;
	export const SYSTEMROOT: string;
	export const TEMP: string;
	export const TERM: string;
	export const TMP: string;
	export const USERDOMAIN: string;
	export const USERDOMAIN_ROAMINGPROFILE: string;
	export const USERNAME: string;
	export const USERPROFILE: string;
	export const VSCODE_CODE_CACHE_PATH: string;
	export const VSCODE_CRASH_REPORTER_PROCESS_TYPE: string;
	export const VSCODE_CWD: string;
	export const VSCODE_ESM_ENTRYPOINT: string;
	export const VSCODE_HANDLES_UNCAUGHT_ERRORS: string;
	export const VSCODE_IPC_HOOK: string;
	export const VSCODE_NLS_CONFIG: string;
	export const VSCODE_PID: string;
	export const WINDIR: string;
	export const ZES_ENABLE_SYSMAN: string;
	export const __PSLockDownPolicy: string;
}

/**
 * Similar to [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private), except that it only includes environment variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Values are replaced statically at build time.
 * 
 * ```ts
 * import { PUBLIC_BASE_URL } from '$env/static/public';
 * ```
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to runtime environment variables, as defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * This module cannot be imported into client-side code.
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * console.log(env.DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` always includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 */
declare module '$env/dynamic/private' {
	export const env: {
		VITE_API_URL: string;
		ACSetupSvcPort: string;
		ACSvcPort: string;
		AI_AGENT: string;
		ALLUSERSPROFILE: string;
		APPDATA: string;
		APPLICATION_INSIGHTS_NO_STATSBEAT: string;
		ChocolateyInstall: string;
		ChocolateyLastPathUpdate: string;
		ChocolateyToolsLocation: string;
		CHROME_CRASHPAD_PIPE_NAME: string;
		CLAUDECODE: string;
		CLAUDE_AGENT_SDK_VERSION: string;
		CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING: string;
		CLAUDE_CODE_ENTRYPOINT: string;
		CLAUDE_CODE_EXECPATH: string;
		CLAUDE_CODE_MAX_OUTPUT_TOKENS: string;
		COMMONPROGRAMFILES: string;
		CommonProgramW6432: string;
		COMPUTERNAME: string;
		COMSPEC: string;
		COPILOT_OTEL_ENABLED: string;
		COPILOT_OTEL_EXPORTER_TYPE: string;
		COPILOT_OTEL_FILE_EXPORTER_PATH: string;
		COREPACK_ENABLE_AUTO_PIN: string;
		DriverData: string;
		EFC_10092_1262719628: string;
		EFC_10092_1592913036: string;
		EFC_10092_2283032206: string;
		EFC_10092_3789132940: string;
		ELECTRON_RUN_AS_NODE: string;
		EnableLog: string;
		EXEPATH: string;
		GIT_EDITOR: string;
		GIT_LFS_PATH: string;
		GOPATH: string;
		GOROOT: string;
		HOME: string;
		HOMEDRIVE: string;
		HOMEPATH: string;
		IGCCSVC_DB: string;
		INIT_CWD: string;
		JAVA_HOME: string;
		LM_STUDIO_API_BASE: string;
		LM_STUDIO_API_KEY: string;
		LOCALAPPDATA: string;
		LOGONSERVER: string;
		MCP_CONNECTION_NONBLOCKING: string;
		MSYSTEM: string;
		NODE: string;
		NoDefaultCurrentDirectoryInExePath: string;
		NODE_ENV: string;
		npm_command: string;
		npm_config_auto_install_peers: string;
		npm_config_cache: string;
		npm_config_frozen_lockfile: string;
		npm_config_manage_package_manager_versions: string;
		npm_config_msvs_version: string;
		npm_config_node_gyp: string;
		npm_config_node_linker: string;
		npm_config_python: string;
		npm_config_registry: string;
		npm_config_user_agent: string;
		npm_config_verify_deps_before_run: string;
		npm_config__jsr_registry: string;
		npm_execpath: string;
		npm_lifecycle_event: string;
		npm_lifecycle_script: string;
		npm_node_execpath: string;
		npm_package_json: string;
		npm_package_name: string;
		npm_package_version: string;
		NUMBER_OF_PROCESSORS: string;
		OLDPWD: string;
		OneDrive: string;
		OS: string;
		OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE: string;
		OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT: string;
		PATH: string;
		PATHEXT: string;
		PLINK_PROTOCOL: string;
		pnpm_config_verify_deps_before_run: string;
		PNPM_HOME: string;
		PNPM_SCRIPT_SRC_DIR: string;
		POWERSHELL_DISTRIBUTION_CHANNEL: string;
		PROCESSOR_ARCHITECTURE: string;
		PROCESSOR_IDENTIFIER: string;
		PROCESSOR_LEVEL: string;
		PROCESSOR_REVISION: string;
		ProgramData: string;
		PROGRAMFILES: string;
		ProgramW6432: string;
		PROMPT: string;
		PSModulePath: string;
		PUBLIC: string;
		PWD: string;
		RlsSvcPort: string;
		SESSIONNAME: string;
		SHELL: string;
		SHLVL: string;
		SYSTEMDRIVE: string;
		SYSTEMROOT: string;
		TEMP: string;
		TERM: string;
		TMP: string;
		USERDOMAIN: string;
		USERDOMAIN_ROAMINGPROFILE: string;
		USERNAME: string;
		USERPROFILE: string;
		VSCODE_CODE_CACHE_PATH: string;
		VSCODE_CRASH_REPORTER_PROCESS_TYPE: string;
		VSCODE_CWD: string;
		VSCODE_ESM_ENTRYPOINT: string;
		VSCODE_HANDLES_UNCAUGHT_ERRORS: string;
		VSCODE_IPC_HOOK: string;
		VSCODE_NLS_CONFIG: string;
		VSCODE_PID: string;
		WINDIR: string;
		ZES_ENABLE_SYSMAN: string;
		__PSLockDownPolicy: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * Similar to [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), but only includes variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Note that public dynamic environment variables must all be sent from the server to the client, causing larger network requests — when possible, use `$env/static/public` instead.
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.PUBLIC_DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
