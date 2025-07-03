export interface Torrent {
  id: string;
  name: string;
  category: string;
  sizeMB: number;
  addedDate: string;
  url: string;
  imageURL?: string;
  csfdRating?: number;
  csfdURL?: string;
  createdAt: string;
  updatedAt: string;
  seeds: number;
  leeches: number;
}

export interface TorrentStats {
  id: string;
  torrentID: string;
  seeds: number;
  leeches: number;
  recordedAt: string;
}

export interface TorrentConnection {
  torrents: Torrent[];
  totalCount: number;
  hasNextPage: boolean;
  hasPreviousPage: boolean;
}

export interface Category {
  name: string;
  count: number;
}

export interface DatabaseStats {
  totalTorrents: number;
  totalCategories: number;
  categoryCounts: Category[];
}

export const TorrentSortBy = {
  NEWEST: 'NEWEST',
  OLDEST: 'OLDEST',
  NAME_ASC: 'NAME_ASC',
  NAME_DESC: 'NAME_DESC',
  SIZE_ASC: 'SIZE_ASC',
  SIZE_DESC: 'SIZE_DESC',
  SEEDS_DESC: 'SEEDS_DESC',
  LEECHES_DESC: 'LEECHES_DESC',
} as const;

export type TorrentSortBy = (typeof TorrentSortBy)[keyof typeof TorrentSortBy];
