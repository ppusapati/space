<script lang="ts">
  import 'virtual:uno.css';
  import '@p9e.in/samavaya/css';
  import '../app.css';
  import { ErpRootLayout } from '@samavāya/ui';

  let { children } = $props();
</script>

<ErpRootLayout>
  {@render children()}
</ErpRootLayout>
       