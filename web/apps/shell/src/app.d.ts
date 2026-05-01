/// <reference types="@sveltejs/kit" />

import type { User, Tenant } from '@samavāya/stores';

declare global {
  namespace App {
    interface Error {
      code?: string;
      message: string;
      details?: Record<string, unknown>;
    }

    interface Locals {
      user: User | null;
      tenant: Tenant | null;
      sessionId: string | null;
    }

    interface PageData {
      user?: User | null;
      tenant?: Tenant | null;
    }

    interface PageState {
      scrollPosition?: number;
    }

    interface Platform {}
  }
}

export {};
