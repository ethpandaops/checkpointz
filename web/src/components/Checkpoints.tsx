import { useQuery } from '@tanstack/react-query';
import { useState, useMemo } from 'react';
import { ExclaimationTriangleIcon } from '@heroicons/react/20/solid'

import Loading from './Loading';
import { APIBeaconSlots, APIBeaconSlot, APIStatus } from '../types';
import CheckpointsTable from './CheckpointsTable';
import SlotSlideout from './SlotSlideout';

export default function Checkpoints() {
  const [slot, setSlot] = useState<APIBeaconSlot | undefined>(undefined);
  const {data, isLoading, error} = useQuery<APIBeaconSlots, Error>(['beacon_slots'], async () => {
		const res = await fetch('/checkpointz/v1/beacon/slots');
		return res.json();
  }, { refetchInterval: 60_000 });
  const { data: statusData } = useQuery<APIStatus, Error>(['status'], async () => {
		const res = await fetch('/checkpointz/v1/status');
		return res.json();
  }, { refetchInterval: 60_000 });
  const latestEpoch = useMemo(() => {
    const finalizedEpoch = statusData?.data?.finality?.finalized?.epoch;
    if (!finalizedEpoch) return;
    return parseInt(finalizedEpoch);
  }, [statusData]);

  if (isLoading) return <div className="flex justify-center pt-10"><Loading /></div>;
  if (!data && error) return (
    <div className="flex justify-center">
      <div className="flex justify-center bg-white/20 py-5 px-10 font-semibold rounded-lg text-gray-100">
        <ExclaimationTriangleIcon className="h-6 w-6 text-yellow-400 pr-1" aria-hidden="true" />Something went wrong
      </div>
    </div>
  );
  return (
    <>
      <SlotSlideout slot={slot?.slot} onClose={() => setSlot(undefined)} />
      <CheckpointsTable
        slots={Object.values(data?.data?.slots ?? {})}
        latestEpoch={latestEpoch}
        onSlotClick={setSlot}
      />
    </>
  )
}
