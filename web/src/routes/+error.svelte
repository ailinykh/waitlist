<script lang="ts">
	import { page } from '$app/state';
</script>

<h1 class="font-bold text-xl mt-8 mb-4">{page.status}: {page?.error?.message}</h1>

<div class="my-8">
	<a href="/" class="font-bold text-sky-600"><span aria-hidden="true">«</span> Back to home</a>
</div>
