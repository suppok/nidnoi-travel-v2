export type ListResponse<T> = {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
};

export type Paging = {
  page: number;
  pageSize: number;
};
