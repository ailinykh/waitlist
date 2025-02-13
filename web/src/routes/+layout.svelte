<script lang="ts">
	import type { Snippet } from 'svelte';
	import '../styles.css';

	let { children }: { children: Snippet } = $props();
</script>

<svelte:head>
	<title>Waitlist</title>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<meta name="description" content="" />
</svelte:head>

<main class="max-w-4xl m-auto">
	{@render children()}
</main>
