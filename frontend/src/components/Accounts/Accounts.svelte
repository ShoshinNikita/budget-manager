<script lang="ts">
	import { onMount } from "svelte";
	import Button, { ButtonSize } from "@src/components/Button.svelte";
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
	const reverseShowClosedAccounts = () => {
		showClosedAccounts = !showClosedAccounts;
	};

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
			{#if showClosedAccounts}
				<Button
					icon={"chevron-up"}
					size={ButtonSize.Medium}
					title="Hide Closed Accounts"
					onClick={reverseShowClosedAccounts}
				/>
			{:else}
				<Button
					icon={"chevron-down"}
					size={ButtonSize.Medium}
					title="Show Closed Accounts"
					onClick={reverseShowClosedAccounts}
				/>
			{/if}
		</div>
	{/if}

	{#if showClosedAccounts}
		{#each closedAccounts as account}
			<Account {account} />
		{/each}
	{/if}

	<div class="actions">
		<Button icon="plus" size={ButtonSize.Large} title="Add" />
		<Button icon="transfer" size={ButtonSize.Large} title="Transfer" />
	</div>
</div>

<style lang="scss">
	.accounts {
		height: 100%;
		position: relative;
	}

	.show-closed-accounts-button {
		margin-top: 20px;
		text-align: center;
	}

	.actions {
		display: grid;
		grid-template-columns: 1fr 1fr;
		position: absolute;
		column-gap: 20px;
		bottom: 50px;
		left: 50%;
		transform: translateX(-50%);
		margin-top: 50px;
	}
</style>
