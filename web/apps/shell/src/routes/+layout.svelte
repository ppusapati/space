<script lang="ts">
  import 'virtual:uno.css';
  import '@@chetana/css';
  import '../app.css';
  import { ErpRootLayout } from '@chetana/ui';

  let { children } = $props();
</script>

<ErpRootLayout>
  {@render children()}
</ErpRootLayout>
       