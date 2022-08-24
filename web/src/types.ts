export interface APIStatus {
  data: {
    upstreams: Record<string, APIUpstream>;
    finality?: APICheckpoints;
    public_url?: string;
    brand_name?: string;
    brand_image_url?: string;
  };
}

export interface APICheckpoints {
  finalized?: APICheckpoint;
  current_justified?: APICheckpoint;
  previous_justified?: APICheckpoint;
}

export interface APICheckpoint {
  epoch: string;
  root: string;
}

export interface APIUpstream {
  name: string;
  healthy: boolean;
  finality?: APICheckpoints;
}

export interface APIBeaconSlot {
  slot: number;
  block_root?: string;
  state_root?: string;
  epoch?: number;
}

export interface APIBeaconSlots {
  data: {
    slots: APIBeaconSlot[];
  };
}

export interface APIBeaconBlockMessageBody {
  randao_reveal?: string;
  execution_payload?: {
    block_hash?: string;
    block_number?: string;
  };
  graffiti?: string;
}

export interface APIBeaconBlockMessage {
  message?: {
    parent_root?: string;
    proposer_index?: string;
    slot?: string;
    state_root?: string;
    body?: APIBeaconBlockMessageBody;
  };
  signature?: string;
}

export interface APIBeaconBlock {
  Version: 'BELLATRIX' | 'ALTAIR' | 'PHASE0';
  Altair?: APIBeaconBlockMessage;
  Bellatrix?: APIBeaconBlockMessage;
  Phase0?: APIBeaconBlockMessage;
}

export interface APIBeaconSlotBlock {
  data: {
    block?: APIBeaconBlock
    epoch?: number;
  }
}
