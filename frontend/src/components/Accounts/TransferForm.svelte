<script lang="ts">
	import Button, { Size as ButtonSize } from "@src/components/Button.svelte";
	import type * as types from "@src/types";
	import { accountService } from "@src/services";

	let accounts: types.AccountWithBalance[] = [];
	let accountsByID = new Map<string, types.AccountWithBalance>();
	accountService.getOpenAccounts((accs) => {
		accountsByID.clear();
		for (const acc of accs) {
			accountsByID.set(acc.id, acc);
		}

		accounts = accs;
	});

	let fromAmount = 0;
	let fromAccountID = "";
	let toAmount = 0;
	let toAccountID = "";

	const getCurrency = (accountID: string) => {
		if (!accountID) {
			return "?";
		}
		const acc = accountsByID.get(accountID);
		if (!acc) {
			return "?";
		}
		return acc.currency;
	};
	$: fromCurrency = getCurrency(fromAccountID);
	$: toCurrency = getCurrency(toAccountID);

	const transfer = () => {
		console.log("TODO: transfer");

		fromAmount = 0;
		fromAccountID = "";
		toAmount = 0;
		toAccountID = "";
	};
</script>

<div class="transfer-window">
	<div class="from">
		<span class="amount">
			<input type="number" bind:value={fromAmount} />
			<span title="Currency">{fromCurrency}</span>
		</span>

		<select bind:value={fromAccountID} title="From">
			<option value="" disabled selected>From Account</option>
			{#each accounts as account (account.id)}
				<option value={account.id}>{account.name}</option>
			{/each}
		</select>
	</div>

	<div class="transfer-button">
		<Button icon="arrow-right" size={ButtonSize.Medium} title="Transfer" onClick={transfer} />
	</div>

	<div class="to">
		<span class="amount">
			<input type="number" bind:value={toAmount} />
			<span title="Currency">{toCurrency}</span>
		</span>

		<select bind:value={toAccountID} title="To">
			<option value="" disabled selected>To Account</option>

			{#each accounts as account (account.id)}
				<option value={account.id}>{account.name}</option>
			{/each}
		</select>
	</div>
</div>

<style lang="scss">
	.transfer-window {
		column-gap: 20px;
		display: grid;
		grid-template-columns: 1fr auto 1fr;
		margin: auto;
		padding: 30px;
	}

	.from,
	.to {
		column-gap: 10px;
		display: grid;
		grid-template-rows: 1fr 1fr;

		> .amount {
			display: grid;
			grid-template-columns: auto min-content;
			column-gap: 3px;
		}
	}

	.transfer-button {
		margin-top: 15px;
		text-align: center;
	}
</style>
