import type * as types from "@src/types";
import * as api from "@src/api";

export class TransactionService {
	notificationService: types.NotificationService;

	constructor(notificationService: types.NotificationService) {
		this.notificationService = notificationService;
	}

	createTransferTransaction = async (args: types.CreateTransferTransactionArgs): Promise<boolean> => {
		const resp = await api.createTransferTransaction(args);
		if (resp instanceof api.Error) {
			this.notificationService.notify(`couldn't create transfer transaction: ${resp.cause}`);
			return false;
		}

		console.log("ok");

		return true;
	};
}
