import rawLogsStyles from "@/pages/DataLogs/components/RawLogs/index.less";
import RawLogsOperations from "@/pages/DataLogs/components/RawLogsOperations";
import RawLogList from "@/pages/DataLogs/components/RawLogList";
import { useModel } from "@@/plugin-model/useModel";
import { Empty } from "antd";
import TableLogList from "@/pages/DataLogs/components/TableLogList";

type RawLogsProps = {};
const RawLogs = (props: RawLogsProps) => {
  const { logs, activeTableLog } = useModel("dataLogs");

  const logList = logs?.logs || [];
  return (
    <div className={rawLogsStyles.rawLogsMain}>
      <div className={rawLogsStyles.rawLogs}>
        {logList.length > 0 ? (
          <>
            <RawLogsOperations />
            {activeTableLog ? <TableLogList /> : <RawLogList />}
          </>
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={"暂无日志信息"}
          />
        )}
      </div>
    </div>
  );
};
export default RawLogs;
