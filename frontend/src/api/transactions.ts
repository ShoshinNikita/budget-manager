import type * as types from "@src/types";
import { makeRequest, Error } from "./api";

export const createTransferTransaction = async (args: types.CreateTransferTransactionArgs): Promise<void | Error> => {
	const resp = await makeRequest<{}>("/api/transactions/create/transfer", {
		date: getCurrentDate(),
		from_account_id: args.fromAccountID,
		from_amount: String(args.fromAmount),
		to_account_id: args.toAccountID,
		to_amount: String(args.toAmount),
	});

	if (resp instanceof Error) {
		return resp;
	}
};

const getCurrentDate = (): string => {
	return new Date().toISOString().slice(0, 10);
};
