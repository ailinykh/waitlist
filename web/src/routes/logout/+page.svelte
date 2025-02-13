<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(async () => {
		localStorage.token = null;
		goto('/login');
	});
</script>
