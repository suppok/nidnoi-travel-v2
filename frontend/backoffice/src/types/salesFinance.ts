// The caller's finance users: who their confirmed LINE bookings can be sent to (GET/PUT /sales-finances).
export type SalesFinanceUser = { id: string; name: string };

export type SalesFinanceSettings = {
  options: SalesFinanceUser[]; // every active finance user except yourself
  financeUserIds: string[];
  defaultFinanceUserId: string; // '' = none
};

export type SalesFinanceUpdateRequest = { financeUserIds: string[]; defaultFinanceUserId: string };
