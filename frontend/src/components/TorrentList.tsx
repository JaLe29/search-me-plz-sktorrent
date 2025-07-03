import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Select, Input, Space, Alert, Row, Col, Button } from 'antd';
import { SearchOutlined, FilterOutlined, FileSearchOutlined } from '@ant-design/icons';
import { useQuery, useLazyQuery } from '@apollo/client';
import { TorrentCard } from './TorrentCard';
import { TorrentCardSkeleton } from './TorrentCardSkeleton';
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
            // Remove duplicates by ID
      const uniqueTorrents = data.torrents.torrents.filter((torrent: Torrent, index: number, self: Torrent[]) =>
        index === self.findIndex((t: Torrent) => t.id === torrent.id)
      );
      setAllTorrents(uniqueTorrents);
      setHasMore(data.torrents.hasNextPage);
    }
  }, [data]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasMore) return;

    setIsLoadingMore(true);

    try {
      const result = await loadMoreQuery({
        variables: {
          first: pageSize,
          after: allTorrents.length.toString(),
          category: selectedCategory || undefined,
          search: searchQuery || undefined,
          sortBy,
        },
      });

      if (result.data?.torrents) {
        const newTorrents = result.data.torrents.torrents;
        setAllTorrents(prev => {
          // Remove duplicates by ID
          const allTorrents = [...prev, ...newTorrents];
          const uniqueTorrents = allTorrents.filter((torrent: Torrent, index: number, self: Torrent[]) =>
            index === self.findIndex((t: Torrent) => t.id === torrent.id)
          );

          return uniqueTorrents;
        });
        setHasMore(result.data.torrents.hasNextPage);
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
        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.08)',
        borderRadius: 16,
        padding: '32px',
        marginBottom: 40,
        border: 'none',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{
              width: 32,
              height: 32,
              background: '#1890ff',
              borderRadius: 8,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}>
              <FilterOutlined style={{ color: 'white', fontSize: 16 }} />
            </div>
            <span style={{ color: '#262626', fontWeight: 600, fontSize: 18 }}>Filtry a vyhledávání</span>
          </div>

          {data?.torrents?.totalCount !== undefined && (
            <div style={{
              background: '#f8f9fa',
              padding: '8px 16px',
              borderRadius: 8,
              border: '1px solid #e2e8f0',
            }}>
              <span style={{ color: '#8c8c8c', fontSize: 14 }}>
                Celkem: <strong style={{ color: '#262626' }}>{data.torrents.totalCount.toLocaleString('cs-CZ')}</strong> torrentů
              </span>
            </div>
          )}
        </div>

        <Space wrap size="large" style={{ width: '100%' }}>
          <div style={{ display: 'flex', width: 320 }}>
            <Input
              placeholder="Hledat torrenty..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onPressEnter={handleSearch}
              style={{
                borderTopRightRadius: 0,
                borderBottomRightRadius: 0,
                borderRight: 'none',
                border: 'none',
                boxShadow: 'none',
                background: '#f8f9fa',
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
                background: '#1890ff',
                border: 'none',
                boxShadow: 'none',
              }}
            >
              Hledat
            </Button>
          </div>

          <Select
            placeholder="Kategorie"
            allowClear
            style={{
              width: 220,
              borderRadius: 12,
            }}
            onChange={handleCategoryChange}
            value={selectedCategory || undefined}
            size="large"
          >
            {categoriesData?.categories?.map((category: Category, index: number) => (
              <Option key={`${category.name}-${index}`} value={category.name}>
                {category.name} ({category.count})
              </Option>
            ))}
          </Select>

          <Select
            placeholder="Řazení"
            style={{
              width: 180,
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

      {/* Results info */}
      {allTorrents.length > 0 && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 24,
          padding: '16px 24px',
          background: 'white',
          borderRadius: 12,
          boxShadow: '0 2px 8px rgba(0, 0, 0, 0.04)',
          border: '1px solid #f0f0f0',
        }}>
          <span style={{ color: '#8c8c8c', fontSize: 14 }}>
            Zobrazeno <strong style={{ color: '#262626' }}>{allTorrents.length}</strong> z <strong style={{ color: '#262626' }}>{data?.torrents?.totalCount?.toLocaleString('cs-CZ') || '?'}</strong> torrentů
          </span>

          {data?.torrents?.totalCount && allTorrents.length < data.torrents.totalCount && (
            <span style={{ color: '#1890ff', fontSize: 14, fontWeight: 500 }}>
              Scrollujte dolů pro další
            </span>
          )}
        </div>
      )}

      {/* Torrents Grid */}
      <Row gutter={[24, 24]} style={{ width: '100%' }} className="torrent-grid">
        {loading ? (
          Array.from({ length: pageSize }).map((_, index) => (
            <Col
              key={`skeleton-${index}`}
              xs={24}
              sm={24}
              md={12}
              lg={12}
              xl={12}
              xxl={12}
              className="torrent-col"
            >
              <TorrentCardSkeleton count={1} />
            </Col>
          ))
        ) : (
          allTorrents.map((torrent: Torrent) => (
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
          ))
        )}
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
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 16 }}>
              <div style={{
                width: 40,
                height: 40,
                border: '3px solid #f0f0f0',
                borderTop: '3px solid #1890ff',
                borderRadius: '50%',
                animation: 'spin 1s linear infinite',
              }} />
              <div style={{ color: '#8c8c8c', fontSize: 14 }}>
                Načítání dalších torrentů...
              </div>
            </div>
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
          padding: '80px 40px',
          background: 'white',
          borderRadius: 16,
          boxShadow: '0 4px 24px rgba(0, 0, 0, 0.06)',
          margin: '40px 0',
        }}>
          <div style={{
            width: 80,
            height: 80,
            background: '#f8f9fa',
            borderRadius: '50%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto 24px',
          }}>
            <FileSearchOutlined style={{ fontSize: 32, color: '#8c8c8c' }} />
          </div>
          <h3 style={{
            color: '#262626',
            fontSize: 20,
            fontWeight: 600,
            margin: '0 0 12px',
          }}>
            Žádné torrenty nebyly nalezeny
          </h3>
          <p style={{
            color: '#8c8c8c',
            fontSize: 16,
            margin: '0 0 24px',
            lineHeight: 1.5,
          }}>
            {searchQuery
              ? `Pro vyhledávání "${searchQuery}" nebyly nalezeny žádné výsledky.`
              : 'Zkuste změnit filtry nebo vyhledávací dotaz.'
            }
          </p>
          <Button
            type="primary"
            size="large"
            onClick={() => {
              setSearchQuery('');
              setSelectedCategory('');
              setSortBy('NEWEST');
            }}
            style={{
              background: '#1890ff',
              border: 'none',
              borderRadius: 8,
            }}
          >
            Zobrazit všechny torrenty
          </Button>
        </div>
      )}


    </div>
  );
};
