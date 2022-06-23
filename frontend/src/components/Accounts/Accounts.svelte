<script lang="ts">
	import { onMount } from "svelte";
	import * as accountStore from "@src/stores/accounts";
	import type * as types from "@src/types";
	import Account from "./Account.svelte";

	let openAccounts: types.AccountWithBalance[] = [];
	accountStore.subscribeToOpenAccountUpdates((accs) => {
		openAccounts = accs;
	});

	let closedAccounts: types.AccountWithBalance[] = [];
	accountStore.subscribeToClosedAccountUpdates((accs) => {
		closedAccounts = accs;
	});

	let showClosedAccounts = false;

	// Init on creation
	onMount(async () => {
		await accountStore.fetchAccounts();
	});
</script>

<div class="card accounts">
	<h2 class="card-title">Accounts</h2>

	{#each openAccounts as account}
		<Account {account} />
	{:else}
		<span>No Accounts Yet</span>
	{/each}

	{#if closedAccounts.length > 0}
		<div class="show-closed-accounts-button">
			<button
				on:click={() => {
					showClosedAccounts = !showClosedAccounts;
				}}
			>
				{#if showClosedAccounts}
					Hide Closed Accs
				{:else}
					Show Closed Accs
				{/if}
			</button>
		</div>
	{/if}

	{#if showClosedAccounts}
		{#each closedAccounts as account}
			<Account {account} />
		{/each}
	{/if}

	<div class="actions">
		<button>Add</button>
		<button>Transfer</button>
	</div>
</div>

<style>
	.accounts {
		height: 100%;
		position: relative;
	}

	.show-closed-accounts-button {
		margin-top: 20px;
		text-align: center;
	}

	.actions {
		text-align: center;
		margin-top: 50px;
	}
</style>
