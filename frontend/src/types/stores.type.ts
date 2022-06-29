import type { Writable } from "svelte/store";
import type { AccountWithBalance } from "./account.type";

export interface AccountStore {
	openAccountStore: Writable<AccountWithBalance[]>;
	closedAccountStore: Writable<AccountWithBalance[]>;
}
