import React, { useState, useCallback } from 'react';
import { FilterContext, defaultFilters, type FilterState } from './FilterContextTypes';

export const FilterProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [filters, setFilters] = useState<FilterState>(defaultFilters);

  const resetFilters = useCallback(() => {
    setFilters(defaultFilters);
  }, []);

  const updateFilters = useCallback((updates: Partial<FilterState>) => {
    setFilters(prev => ({ ...prev, ...updates }));
  }, []);

  return (
    <FilterContext.Provider value={{ filters, resetFilters, updateFilters }}>
      {children}
    </FilterContext.Provider>
  );
};