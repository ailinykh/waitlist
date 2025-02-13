<script lang="ts">
	import { onMount } from 'svelte';

	let username: string | null = null;

	onMount(async () => {
		const data = await fetch('/api/telegram/oauth').then((res) => res.json());
		username = data.username;
	});
</script>

<div class="text-center">
	<h1 class="font-bold text-xl mt-8 mb-4">Waitlist</h1>
	<p class="my-8">At this moment only Telegram Login is available</p>
	{#if username}
		<div class="flex justify-center">
			<script
				async
				src="https://telegram.org/js/telegram-widget.js?22"
				data-telegram-login={username}
				data-size="large"
				data-onauth="onTelegramAuth(user)"
				data-request-access="write"
			>
			</script>
			<script type="text/javascript">
				async function onTelegramAuth(user) {
					const u = new URLSearchParams(user).toString();

					const data = await fetch(`/api/telegram/oauth/token?${u}`).then((res) => res.json());
					localStorage.token = data.token;
					location.href = '/';
				}
			</script>
		</div>
	{/if}
</div>
