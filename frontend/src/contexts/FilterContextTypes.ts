import { createContext } from 'react';
import type { TorrentSortBy } from '../types/graphql';

export interface FilterState {
  searchQuery: string;
  csfdIDQuery: string;
  selectedCategory: string;
  sortBy: TorrentSortBy;
  isSearchingByCSFD: boolean;
  isSearchActive: boolean;
}

export interface FilterContextType {
  filters: FilterState;
  resetFilters: () => void;
  updateFilters: (updates: Partial<FilterState>) => void;
}

export const defaultFilters: FilterState = {
  searchQuery: '',
  csfdIDQuery: '',
  selectedCategory: '',
  sortBy: 'NEWEST',
  isSearchingByCSFD: false,
  isSearchActive: false,
};

export const FilterContext = createContext<FilterContextType | undefined>(undefined);