<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let token = localStorage.token;
	let entries: [App.Entry] | null = null;

	onMount(async () => {
		if (!token) {
			console.log('token not found, redirecting to /login page');
			goto('/login');
		} else {
			entries = await fetch('/api/entries', {
				headers: { Authorization: `Bearer ${token}` }
			}).then((res) => res.json()); // TODO: error handle
		}
	});
</script>

{#if entries}
	<div>
		<h1 class="font-bold text-xl mt-8 mb-4">Waitlist</h1>
		<div class="flex justify-between my-8">
			<p>We have {entries.length} entries for now.</p>
			<a href="/logout" class="font-bold text-sky-600">Logout <span aria-hidden="true">»</span></a>
		</div>
		<table class="table-auto w-full text-sm">
			<thead>
				<tr class="border-b-2 border-stone-200">
					<th class="p-2 text-left" scope="col">Bot</th>
					<th class="p-2 text-left" scope="col">User ID</th>
					<th class="p-2 text-left" scope="col">FirstName</th>
					<th class="p-2 text-left" scope="col">LastName</th>
					<th class="p-2 text-left" scope="col">Username</th>
					<th class="p-2 text-left" scope="col">Message</th>
					<th class="p-2 text-left" scope="col">Timestamp</th>
				</tr>
			</thead>
			<tbody>
				{#each entries as e}
					<tr class="border-t-1 border-stone-200">
						<td class="p-2">{e.bot_username}</td>
						<td class="p-2">{e.user_id}</td>
						<td class="p-2">{e.first_name}</td>
						<td class="p-2">{e.last_name}</td>
						<td class="p-2">{e.username}</td>
						<td class="p-2">{e.message}</td>
						<td class="p-2">{new Date(Date.parse(e.created_at)).toLocaleString('ru-RU')}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
