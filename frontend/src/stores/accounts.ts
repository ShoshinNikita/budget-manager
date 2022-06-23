import { writable } from "svelte/store";
import { getAccounts } from "@src/api";
import * as types from "@src/types";

const openAccountStore = writable<types.AccountWithBalance[]>([]);
const closedAccountStore = writable<types.AccountWithBalance[]>([]);

export const fetchAccounts = async () => {
	const accounts = await getAccounts();
	if (!accounts) {
		return;
	}

	let openAccounts: types.AccountWithBalance[] = [];
	let closedAccounts: types.AccountWithBalance[] = [];
	for (const acc of accounts) {
		if (acc.status == types.AccountStatus.Open) {
			openAccounts.push(acc);
		} else {
			closedAccounts.push(acc);
		}
	}

	openAccountStore.set(openAccounts);
	closedAccountStore.set(closedAccounts);
};

export const subscribeToOpenAccountUpdates = openAccountStore.subscribe;
export const subscribeToClosedAccountUpdates = closedAccountStore.subscribe;
