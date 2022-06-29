import { writable } from "svelte/store";
import type * as types from "@src/types";

export class AccountStore {
	openAccountStore = writable<types.AccountWithBalance[]>([]);
	closedAccountStore = writable<types.AccountWithBalance[]>([]);
}
