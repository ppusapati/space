<script lang="ts">
  let { children } = $props();
</script>
 
<div class="grid min-h-screen">
  <div class="flex flex-col justify-center p-xl max-w-[480px] w-full mx-auto">
    <div class="text-center mb-2xl pt-xl">
      <h1 class="text-[2rem] font-bold text-primary mb-xs">samavāya</h1>
      <p class="text-sm text-text-secondary">Enterprise Resource Planning</p>
    </div>

    <div class="flex-1 flex flex-col justify-center">
      {@render children()}
    </div>

    <footer class="text-center pt-xl">
      <p class="text-xs text-text-muted">&copy; {new Date().getFullYear()} samavāya ERP. All rights reserved.</p>
    </footer>
  </div>

  <div class="hidden lg:block flex-1 bg-gradient-to-br from-primary to-primary-hover relative overflow-hidden">
    <div class="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg width=%2760%27 height=%2760%27 viewBox=%270 0 60 60%27 xmlns=%27http://www.w3.org/2000/svg%27%3E%3Cg fill=%27none%27 fill-rule=%27evenodd%27%3E%3Cg fill=%27%23ffffff%27 fill-opacity=%270.05%27%3E%3Cpath d=%27M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z%27/%3E%3C/g%3E%3C/g%3E%3C/svg%3E')]"></div>
  </div>
</div>
