<script lang="ts">
	import * as types from "@src/types";
	import Button from "@src/components/Button.svelte";
	import { accountService } from "@src/services";

	export let account: types.AccountWithBalance | undefined = undefined;
	export let isNewAccountForm = false;

	// For new account
	let newAccountName = "",
		newAccountCurreny = "USD";

	const currencies = ["USD", "RUB", "EUR"];

	const createAccount = async () => {
		const ok = await accountService.createAccount(newAccountName, newAccountCurreny);
		if (ok) {
			newAccountName = "";
			newAccountCurreny = "USD";
		}
	};

	// For existing account
	let isAccountOpen = false,
		accountName = "",
		isEditMode = false;
	if (account) {
		isAccountOpen = account.status === types.AccountStatus.Open;
		accountName = account.name;
	}

	const editAccount = async () => {
		isEditMode = false;

		await accountService.editAccount(account!.id, accountName);
	};

	const closeAccount = async () => {
		await accountService.closeAccount(account!);
	};

	const resetChanges = () => {
		isEditMode = false;
		accountName = account!.name;
	};

	// Utils
	const onEnter = (f: () => void) => (ev: KeyboardEvent) => {
		if (ev.key === "Enter") {
			f();
		}
	};
</script>

<tr class="account">
	<td class="name">
		{#if isNewAccountForm}
			<input
				type="text"
				bind:value={newAccountName}
				on:keypress={onEnter(createAccount)}
				placeholder="{newAccountCurreny} account"
			/>
		{:else}
			<input type="text" bind:value={accountName} disabled={!isEditMode} on:keypress={onEnter(editAccount)} />
		{/if}
	</td>
	<td class="balance">
		{#if isNewAccountForm}
			0
			<select bind:value={newAccountCurreny}>
				{#each currencies as currency}
					<option value={currency} selected={currency === newAccountCurreny}>{currency}</option>
				{/each}
			</select>
		{:else}
			{account?.balance}
			{account?.currency}
		{/if}
	</td>
	<td class="actions">
		{#if isNewAccountForm}
			<Button icon="check" title="Create" onClick={createAccount} />
		{:else}
			{#if !isEditMode}
				<Button
					icon="edit-2"
					disabled={!isAccountOpen}
					title="Edit"
					onClick={() => {
						isEditMode = true;
					}}
				/>
			{:else}
				<Button icon="check" title="Save" onClick={editAccount} />
			{/if}

			{#if !isEditMode}
				<Button icon="x" disabled={!isAccountOpen || isEditMode} title="Close" onClick={closeAccount} />
			{:else}
				<Button icon="rotate-ccw" title="Reset" onClick={resetChanges} />
			{/if}
		{/if}
	</td>
</tr>

<style lang="scss">
	td {
		padding-right: 15px;

		&:last-child {
			padding-right: 0;
		}
	}

	input:disabled {
		border-color: rgba($color: #000000, $alpha: 0);
		color: #000000;
	}

	.name {
		min-width: 100px;

		> input {
			text-overflow: ellipsis;
		}
	}

	.balance,
	.actions {
		width: 1%;
		white-space: nowrap;
	}
</style>
