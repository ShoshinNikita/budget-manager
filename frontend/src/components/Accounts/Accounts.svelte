<script lang="ts">
	import type * as types from "@src/types/account.type";
	import Account from "./Account.svelte";

	// TODO: use store
	export let accounts: types.AccountWithBalance[] = [];
</script>

<div class="card accounts">
	<h2 class="card-title">Accounts</h2>

	{#each accounts as account}
		<Account {account} />
	{:else}
		<span>No Accounts Yet</span>
	{/each}

	<div class="actions">
		<button>Add New</button>
		<button>Transfer</button>
	</div>
</div>

<style>
	.accounts {
		height: 100%;
		position: relative;
	}

	.actions {
		text-align: center;
		margin-top: 30px;
	}
</style>
