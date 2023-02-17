import { ExclamationTriangleIcon } from '@heroicons/react/20/solid';

import Loading from '@components/Loading';
import useStatus from '@hooks/status';

import UpstreamTable from './UpstreamTable';

export default function Upstream() {
  const { data, isLoading, error } = useStatus({ refetchInterval: 60_000 });
  if (isLoading)
    return (
      <div className="flex justify-center pt-10">
        <Loading />
      </div>
    );
  if (!data && error)
    return (
      <div className="flex justify-center">
        <div className="flex justify-center bg-white/20 mt-5 py-5 px-10 font-semibold rounded-lg text-gray-100">
          <ExclamationTriangleIcon className="h-6 w-6 text-yellow-400 pr-1" aria-hidden="true" />
          Something went wrong
        </div>
      </div>
    );
  return <UpstreamTable upstreams={Object.values(data?.data?.upstreams ?? {})} />;
}
