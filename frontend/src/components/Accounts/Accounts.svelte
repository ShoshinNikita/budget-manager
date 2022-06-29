<script lang="ts">
	import Button, { ButtonSize } from "@src/components/Button.svelte";
	import type * as types from "@src/types";
	import { accountService } from "@src/services";
	import Account from "./Account.svelte";

	let openAccounts: types.AccountWithBalance[] = [];
	accountService.getOpenAccounts((accs) => {
		openAccounts = accs;
	});

	let closedAccounts: types.AccountWithBalance[] = [];
	accountService.getClosedAccounts((accs) => {
		closedAccounts = accs;
	});

	let showClosedAccounts = false;
	const reverseShowClosedAccounts = () => {
		showClosedAccounts = !showClosedAccounts;
	};
</script>

<div class="card accounts">
	<h2 class="card-title">Accounts</h2>

	<table class="accounts-list">
		{#each openAccounts as account (account.id)}
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
			{#each closedAccounts as account (account.id)}
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
