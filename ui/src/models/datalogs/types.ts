import {
  HighChartsResponse,
  LogsResponse,
  TablesResponse,
} from "@/services/dataLogs";

export interface QueryParams {
  logLibrary?: TablesResponse;
  page?: number;
  pageSize?: number;
  st?: number;
  et?: number;
  kw?: string;
}

export type PaneType = {
  pane: string;
  paneId: string;
  paneType: number;
  start?: number;
  end?: number;
  keyword?: string;
  activeTabKey?: string;
  activeIndex?: number;
  page?: number;
  pageSize?: number;
  logs: LogsResponse | undefined;
  highCharts: HighChartsResponse | undefined;
};

export interface Extra {
  isPaging?: boolean; // 是否是分页
  reqParams?: QueryParams; // 请求参数
}
