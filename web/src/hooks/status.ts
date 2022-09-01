import { useQuery, UseQueryOptions } from '@tanstack/react-query';

import { APIStatus } from '@types';

export default function useStatus(options?: UseQueryOptions<APIStatus, Error>) {
  return useQuery<APIStatus, Error>(
    ['status'],
    async () => {
      const res = await fetch('/checkpointz/v1/status');
      return res.json();
    },
    options,
  );
}
