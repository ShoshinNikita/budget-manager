import * as types from "@src/types";
import { makeRequest, Error } from "./api";

export const getAccounts = async (): Promise<types.AccountWithBalance[] | Error> => {
	const resp = await makeRequest<{
		accounts: {
			id: string;
			name: string;
			balance: string;
			currency: string;
			status: string;
			created_at: string;
			updated_at: string;
		}[];
	}>("/api/accounts/get", {});

	if (resp instanceof Error) {
		return resp;
	}

	let res: types.AccountWithBalance[] = [];
	for (let account of resp.accounts) {
		let status = types.AccountStatus.Close;
		if (account.status === "open") {
			status = types.AccountStatus.Open;
		}

		res.push({
			id: account.id,
			name: account.name,
			balance: account.balance,
			currency: account.currency,
			status: status,
			createdAt: new Date(account.created_at),
			updatedAt: new Date(account.updated_at),
		});
	}

	return res;
};

export const createAccount = async (name: string, currency: string): Promise<void | Error> => {
	const resp = await makeRequest<{}>("/api/accounts/create", {
		name: name,
		currency: currency,
	});

	if (resp instanceof Error) {
		return resp;
	}
};

export const editAccount = async (): Promise<void | Error> => {
	return new Error("not implemented yet");
};

export const closeAccount = async (id: string): Promise<void | Error> => {
	const resp = await makeRequest<{}>("/api/accounts/close", { id: id });
	if (resp instanceof Error) {
		return resp;
	}
};
