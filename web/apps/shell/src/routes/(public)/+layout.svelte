<!--
  Public-route layout. Used by /login, /login/webauthn,
  /reset-password.

  Renders a centred 480px column with the chetana brand mark + a
  small footer. The (auth) layout already exists for the legacy
  identity flows; (public) is the chetana-platform-flavoured
  twin so the new auth pages stay scoped to the platform brand.
-->
<script lang="ts">
  let { children } = $props();
</script>

<div class="grid min-h-screen bg-background">
  <div class="flex flex-col justify-center p-xl max-w-[480px] w-full mx-auto">
    <div class="text-center mb-2xl pt-xl">
      <h1 class="text-[2rem] font-bold text-primary mb-xs">Chetana</h1>
      <p class="text-sm text-text-secondary">Space platform</p>
    </div>
    <div class="flex-1 flex flex-col justify-center">
      {@render children()}
    </div>
    <footer class="text-center pt-xl">
      <p class="text-xs text-text-muted">
        &copy; {new Date().getFullYear()} Chetana. All rights reserved.
      </p>
    </footer>
  </div>
</div>
