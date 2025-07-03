import React from 'react';
import { Tag, Space, Typography, Image, Tooltip, Button } from 'antd';
import {
  DownloadOutlined,
  EyeOutlined,
  CalendarOutlined,
} from '@ant-design/icons';
import type { Torrent } from '../types/graphql';
import './TorrentCard.css';

const { Text, Title } = Typography;

interface TorrentCardProps {
  torrent: Torrent;
}

export const TorrentCard: React.FC<TorrentCardProps> = ({ torrent }) => {
  const formatSize = (sizeMB: number): string => {
    if (sizeMB >= 1024 * 1024) {
      return `${(sizeMB / (1024 * 1024)).toFixed(1)} TB`;
    }
    if (sizeMB >= 1024) {
      return `${(sizeMB / 1024).toFixed(1)} GB`;
    }
    return `${sizeMB.toFixed(1)} MB`;
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleDateString('cs-CZ');
  };

  const shouldShowRating = (torrent.csfdRating ?? 0) > 0;

  return (
    <div className="torrent-card">
      {/* Gradient overlay */}
      <div className="torrent-card-gradient" />

      {/* Responsive layout container */}
      <div className="torrent-card-content">
        {torrent.imageURL && (
          <div className="torrent-card-image-container">
            <Image
              className="torrent-card-image"
              src={torrent.imageURL}
              alt={torrent.name}
              fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
              preview={false}
            />
            {shouldShowRating && (
              <div className="torrent-card-rating">
                {torrent.csfdRating}%
                {/* Debug: {JSON.stringify({ rating: torrent.csfdRating, shouldShow: shouldShowRating })} */}
              </div>
            )}
          </div>
        )}

        <div className="torrent-card-info">
          <Title level={4} className="torrent-card-title">
            {torrent.name}
          </Title>

          <div className="torrent-card-category">
            <Tag className="torrent-card-tag">
              {torrent.category}
            </Tag>
          </div>

          <div className="torrent-card-date">
            <Text className="torrent-card-date-text">
              <CalendarOutlined style={{ marginRight: 6 }} />
              {formatDate(torrent.addedDate)}
            </Text>
          </div>

          <div className="torrent-card-size">
            <Text className="torrent-card-size-text">
              {formatSize(torrent.sizeMB)}
            </Text>
          </div>

          <div className="torrent-card-stats">
            <Space size="small">
              <Tag className="torrent-card-stat torrent-card-seeds">
                {torrent.seeds} ↑
              </Tag>
              <Tag className="torrent-card-stat torrent-card-leeches">
                {torrent.leeches} ↓
              </Tag>
            </Space>
          </div>
        </div>
      </div>

      {/* Action buttons */}
      <div className="torrent-card-actions">
        <Button
          type="primary"
          size="small"
          icon={<DownloadOutlined />}
          className="torrent-card-download-btn"
          onClick={() => window.open(torrent.url, '_blank')}
        >
          Stáhnout
        </Button>

        {torrent.csfdURL && (
          <Tooltip title="Zobrazit na ČSFD">
            <Button
              size="small"
              icon={<EyeOutlined />}
              className="torrent-card-csfd-btn"
              onClick={() => window.open(torrent.csfdURL, '_blank')}
            />
          </Tooltip>
        )}
      </div>
    </div>
  );
};
