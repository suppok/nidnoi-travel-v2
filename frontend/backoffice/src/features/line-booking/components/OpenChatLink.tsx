import { MessageOutlined } from '@ant-design/icons';
import { Button } from 'antd';

// "เปิดแชท": opens the customer's chat in LINE OA Manager in a new tab. The URL is pasted by sales
// (OA Manager's chat ID isn't the Messaging API user ID); renders nothing until it is saved.
export default function OpenChatLink({ url }: { url: string | undefined }) {
  if (!url) return null;
  return (
    <Button type="link" size="small" icon={<MessageOutlined />} href={url} target="_blank" rel="noopener" className="px-1">
      เปิดแชท
    </Button>
  );
}
