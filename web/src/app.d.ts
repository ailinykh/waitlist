// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		interface Entry {
			bot_username: string;
			user_id: string;
			first_name: string;
			last_name: string;
			username: string;
			message: string;
			created_at: string;
		}
	}
}

export {};
