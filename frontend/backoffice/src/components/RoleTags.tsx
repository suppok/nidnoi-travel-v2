import { Tag } from 'antd';
import { type Role, roleInfo } from '../constants/roles.ts';

export default function RoleTags({ roles }: { roles: readonly Role[] }) {
  return (
    <span className="inline-flex flex-wrap gap-1">
      {roles.map((role) => {
        const info = roleInfo(role);
        return (
          <Tag key={role} color={info.color} className="m-0">
            {info.label}
          </Tag>
        );
      })}
    </span>
  );
}
