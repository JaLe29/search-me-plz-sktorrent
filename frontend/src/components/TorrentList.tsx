import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Select, Input, Space, Spin, Alert, Row, Col, Button } from 'antd';
import { SearchOutlined, FilterOutlined } from '@ant-design/icons';
import { useQuery, useLazyQuery } from '@apollo/client';
import { TorrentCard } from './TorrentCard';
import { GET_TORRENTS, GET_CATEGORIES } from '../graphql/queries';
import type { TorrentSortBy, Category, Torrent } from '../types/graphql';

const { Option } = Select;

interface TorrentListProps {
  pageSize?: number;
}

export const TorrentList: React.FC<TorrentListProps> = ({ pageSize = 24 }) => {
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [sortBy, setSortBy] = useState<TorrentSortBy>('NEWEST');
  const [allTorrents, setAllTorrents] = useState<Torrent[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const observerRef = useRef<IntersectionObserver | null>(null);
  const loadingRef = useRef<HTMLDivElement>(null);

  const { data: categoriesData } = useQuery(GET_CATEGORIES);

  // Initial load
  const { data, loading, error } = useQuery(GET_TORRENTS, {
    variables: {
      first: pageSize,
      after: null,
      category: selectedCategory || undefined,
      search: searchQuery || undefined,
      sortBy,
    },
    fetchPolicy: 'cache-and-network',
  });

  // Lazy query for loading more
  const [loadMoreQuery, { loading: loadingMore }] = useLazyQuery(GET_TORRENTS, {
    fetchPolicy: 'no-cache',
  });

  // Reset data when filters change
  useEffect(() => {
    setAllTorrents([]);
    setHasMore(true);
    setIsLoadingMore(false);
  }, [selectedCategory, searchQuery, sortBy]);

  // Update all torrents when initial data arrives
  useEffect(() => {
    if (data?.torrents?.torrents) {
      setAllTorrents(data.torrents.torrents);
      setHasMore(data.torrents.hasNextPage);
      console.log('Initial load:', data.torrents.torrents.length, 'torrents');
    }
  }, [data]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasMore) return;

    setIsLoadingMore(true);
    const currentLength = allTorrents.length;

    try {
      const result = await loadMoreQuery({
        variables: {
          first: pageSize,
          after: currentLength.toString(),
          category: selectedCategory || undefined,
          search: searchQuery || undefined,
          sortBy,
        },
      });

      if (result.data?.torrents) {
        const newTorrents = result.data.torrents.torrents;
        setAllTorrents(prev => [...prev, ...newTorrents]);
        setHasMore(result.data.torrents.hasNextPage);
        console.log('Loaded more torrents:', newTorrents.length, 'Total:', allTorrents.length + newTorrents.length);
      }
    } catch (error) {
      console.error('Error loading more torrents:', error);
    } finally {
      setIsLoadingMore(false);
    }
  }, [isLoadingMore, hasMore, allTorrents.length, pageSize, selectedCategory, searchQuery, sortBy, loadMoreQuery]);

  // Intersection Observer for infinite scroll
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !isLoadingMore && !loadingMore) {
          console.log('Intersection detected, loading more...');
          loadMore();
        }
      },
      { threshold: 0.1, rootMargin: '300px' }
    );

    if (loadingRef.current) {
      observer.observe(loadingRef.current);
    }

    observerRef.current = observer;

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, [loadMore, hasMore, isLoadingMore, loadingMore]);

  const handleSearch = () => {
    setAllTorrents([]);
    setHasMore(true);
    setIsLoadingMore(false);
  };

  const handleCategoryChange = (value: string) => {
    setSelectedCategory(value);
    setAllTorrents([]);
    setHasMore(true);
    setIsLoadingMore(false);
  };

  const handleSortChange = (value: TorrentSortBy) => {
    setSortBy(value);
    setAllTorrents([]);
    setHasMore(true);
    setIsLoadingMore(false);
  };

  if (error) {
    return (
      <Alert
        message="Chyba při načítání torrentů"
        description={error.message}
        type="error"
        showIcon
        style={{
          borderRadius: 12,
        }}
      />
    );
  }

  return (
    <div>
      {/* Moderní filtry */}
      <div style={{
        background: 'white',
        boxShadow: '0 4px 24px rgba(0, 0, 0, 0.06)',
        border: '1px solid #f0f0f0',
        borderRadius: 20,
        padding: '24px 32px',
        marginBottom: 40,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 16 }}>
          <FilterOutlined style={{ color: '#1890ff', fontSize: 16 }} />
          <span style={{ color: '#262626', fontWeight: 500, fontSize: 14 }}>Filtry</span>
        </div>

        <Space wrap size="large" style={{ width: '100%' }}>
          <div style={{ display: 'flex', width: 300 }}>
            <Input
              placeholder="Hledat torrenty..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onPressEnter={handleSearch}
              style={{
                borderTopRightRadius: 0,
                borderBottomRightRadius: 0,
                borderRight: 'none',
              }}
              prefix={<SearchOutlined style={{ color: '#1890ff' }} />}
              size="large"
            />
            <Button
              type="primary"
              size="large"
              onClick={handleSearch}
              style={{
                borderTopLeftRadius: 0,
                borderBottomLeftRadius: 0,
                background: 'linear-gradient(135deg, #1890ff 0%, #722ed1 100%)',
                border: 'none',
              }}
            >
              Hledat
            </Button>
          </div>

          <Select
            placeholder="Kategorie"
            allowClear
            style={{
              width: 200,
              borderRadius: 12,
            }}
            onChange={handleCategoryChange}
            value={selectedCategory || undefined}
            size="large"
          >
            {categoriesData?.categories?.map((category: Category) => (
              <Option key={category.name} value={category.name}>
                {category.name} ({category.count})
              </Option>
            ))}
          </Select>

          <Select
            placeholder="Řazení"
            style={{
              width: 160,
              borderRadius: 12,
            }}
            onChange={handleSortChange}
            value={sortBy}
            size="large"
          >
            <Option value="NEWEST">Nejnovější</Option>
            <Option value="OLDEST">Nejstarší</Option>
            <Option value="NAME_ASC">Název A-Z</Option>
            <Option value="NAME_DESC">Název Z-A</Option>
            <Option value="SIZE_DESC">Největší</Option>
            <Option value="SIZE_ASC">Nejmenší</Option>
            <Option value="SEEDS_DESC">Nejvíce seedů</Option>
            <Option value="LEECHES_DESC">Nejvíce leecherů</Option>
          </Select>
        </Space>
      </div>

      {/* Torrents Grid */}
      <Row gutter={[24, 24]} style={{ width: '100%' }} className="torrent-grid">
        {allTorrents.map((torrent: Torrent) => (
          <Col
            key={torrent.id}
            xs={24}
            sm={24}
            md={12}
            lg={12}
            xl={12}
            xxl={12}
            className="torrent-col"
          >
            <TorrentCard torrent={torrent} />
          </Col>
        ))}
      </Row>

      {/* Loading indicator for infinite scroll */}
      {hasMore && (
        <div
          ref={loadingRef}
          style={{
            textAlign: 'center',
            padding: '40px',
            marginTop: 20,
          }}
        >
          {isLoadingMore || loadingMore ? (
            <Spin size="large" />
          ) : (
            <div style={{ color: '#8c8c8c', fontSize: 14 }}>
              Scrollujte dolů pro načtení dalších torrentů...
            </div>
          )}
        </div>
      )}

      {/* End of results */}
      {!hasMore && allTorrents.length > 0 && (
        <div style={{
          textAlign: 'center',
          padding: '40px',
          color: '#8c8c8c',
          fontSize: 14,
        }}>
          Zobrazeny všechny výsledky ({allTorrents.length} torrentů)
        </div>
      )}

      {/* Empty state */}
      {!loading && allTorrents.length === 0 && (
        <div style={{
          textAlign: 'center',
          padding: '80px',
          color: '#8c8c8c',
          fontSize: 16,
        }}>
          Žádné torrenty nebyly nalezeny
        </div>
      )}

      {/* Debug info */}
      <div style={{
        position: 'fixed',
        bottom: 10,
        right: 10,
        background: 'rgba(0,0,0,0.8)',
        color: 'white',
        padding: '8px 12px',
        borderRadius: 8,
        fontSize: 12,
        zIndex: 1000,
      }}>
        Debug: {allTorrents.length} torrentů, hasMore: {hasMore.toString()}, loading: {loading.toString()}, loadingMore: {(isLoadingMore || loadingMore).toString()}
      </div>
    </div>
  );
};
