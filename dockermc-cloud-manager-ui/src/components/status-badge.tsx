import type { ContainerStatus } from "@/types/api";
import { Badge } from "./ui/badge";

function getStatusColor(status: ContainerStatus) {
  switch (status) {
    case 'running':
      return 'bg-green-500';
    case 'stopped':
      return 'bg-gray-500';
    case 'creating':
      return 'bg-blue-500';
    case 'error':
      return 'bg-red-500';
    default:
      return 'bg-gray-500';
  }
}

export const StatusBadge = ({ status }: { status: ContainerStatus }) => {
    return (
        <Badge className={getStatusColor(status)}>
          {status}
        </Badge>
    )
}