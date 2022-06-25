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

	<table class="accounts-list">
		{#each openAccounts as account}
			<Account {account} />
		{/each}

		<Account isNewAccountForm={true} />

		{#if closedAccounts.length > 0}
			<tr class="show-closed-accounts-button">
				<td colspan="3">
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
				</td>
			</tr>
		{/if}

		{#if showClosedAccounts}
			{#each closedAccounts as account}
				<Account {account} />
			{/each}
		{/if}
	</table>

	<div class="actions">
		<Button icon="transfer" size={ButtonSize.Large} title="Transfer" />
	</div>
</div>

<style lang="scss">
	.accounts {
		height: 100%;
		position: relative;
	}

	.accounts-list {
		width: 100%;
	}

	.show-closed-accounts-button {
		text-align: center;
	}

	.actions {
		text-align: center;
		position: absolute;
		bottom: 50px;
		left: 50%;
		transform: translateX(-50%);
	}
</style>
