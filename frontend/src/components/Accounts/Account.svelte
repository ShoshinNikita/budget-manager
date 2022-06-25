<script lang="ts">
	import * as types from "@src/types";
	import Button from "@src/components/Button.svelte";
	import * as api from "@src/api";

	export let account: types.AccountWithBalance | undefined = undefined;
	export let isNewAccountForm = false;

	// For new account
	let newAccountName = "",
		newAccountCurreny = "USD";

	// TODO: get from store
	const currencies = ["USD", "RUB", "EUR"];

	const createAccount = () => {
		if (newAccountName === "") {
			return;
		}

		const resp = api.createAccount(newAccountName, newAccountCurreny);
		if (resp instanceof api.Error) {
			handleError(resp);
		}

		// TODO: trigger accounts refresh
	};

	// For existing account
	let isAccountOpen = false,
		accountName = "",
		isEditMode = false;
	if (account) {
		isAccountOpen = account.status === types.AccountStatus.Open;
		accountName = account.name;
	}

	const editAccount = () => {
		isEditMode = false;
		if (accountName === account!.name) {
			return;
		}

		const resp = api.editAccount();
		if (resp instanceof api.Error) {
			handleError(resp);
		}

		// TODO: trigger accounts refresh
	};

	const closeAccount = () => {
		if (!confirm(`Do you really want to close account "${account!.name}"`)) {
			return;
		}

		const resp = api.closeAccount(account!.id);
		if (resp instanceof api.Error) {
			handleError(resp);
		}

		// TODO: trigger accounts refresh
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

	// TODO: show notification
	const handleError = (err: api.Error) => {
		console.log(err);
	};
</script>

<tr class="account">
	<td class="name">
		{#if isNewAccountForm}
			<input type="text" bind:value={newAccountName} on:keypress={onEnter(createAccount)} placeholder="Name" />
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

	// TODO: move to global
	input {
		width: 100%;
		font-size: inherit;
		border: none;
		border-bottom: 1px solid rgba($color: #000000, $alpha: 0.6);
		padding: 0;

		&:disabled {
			border-color: rgba($color: #000000, $alpha: 0);
		}
	}

	select {
		background-color: white;
		border: none;
		border-bottom: 1px solid black;
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
