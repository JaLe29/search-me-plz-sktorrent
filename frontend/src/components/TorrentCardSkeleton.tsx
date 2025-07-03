import React from 'react';
import { Skeleton } from 'antd';
import './TorrentCardSkeleton.css';

interface TorrentCardSkeletonProps {
  count?: number;
}

export const TorrentCardSkeleton: React.FC<TorrentCardSkeletonProps> = ({ count = 1 }) => {
  return (
    <>
      {Array.from({ length: count }).map((_, index) => (
        <div key={index} className="torrent-card-skeleton">
          {/* Gradient overlay */}
          <div className="torrent-card-skeleton-gradient" />

          {/* Content */}
          <div className="torrent-card-skeleton-content">
            {/* Image skeleton */}
            <div className="torrent-card-skeleton-image-container">
              <Skeleton.Image
                active
                className="torrent-card-skeleton-image"
              />
            </div>

            {/* Info skeleton */}
            <div className="torrent-card-skeleton-info">
              {/* Title */}
              <Skeleton
                active
                paragraph={{ rows: 2, width: ['100%', '80%'] }}
                className="torrent-card-skeleton-title"
              />

              {/* Category */}
              <div className="torrent-card-skeleton-category">
                <Skeleton.Button active size="small" style={{ width: 80, height: 24 }} />
              </div>

              {/* Date */}
              <div className="torrent-card-skeleton-date">
                <Skeleton.Input active size="small" style={{ width: 100, height: 16 }} />
              </div>

              {/* Size */}
              <div className="torrent-card-skeleton-size">
                <Skeleton.Input active size="small" style={{ width: 60, height: 20 }} />
              </div>

              {/* Stats */}
              <div className="torrent-card-skeleton-stats">
                <Skeleton.Button active size="small" style={{ width: 60, height: 24 }} />
                <Skeleton.Button active size="small" style={{ width: 60, height: 24 }} />
              </div>
            </div>
          </div>

          {/* Actions skeleton */}
          <div className="torrent-card-skeleton-actions">
            <Skeleton.Button active style={{ width: '100%', height: 32 }} />
            <Skeleton.Button active size="small" style={{ width: 40, height: 32 }} />
          </div>
        </div>
      ))}
    </>
  );
};