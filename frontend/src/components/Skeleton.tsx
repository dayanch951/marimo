import React from 'react';
import './Skeleton.css';

interface SkeletonProps {
  variant?: 'text' | 'circular' | 'rectangular';
  width?: string | number;
  height?: string | number;
  animation?: 'pulse' | 'wave' | 'none';
  count?: number;
}

const Skeleton: React.FC<SkeletonProps> = ({
  variant = 'text',
  width,
  height,
  animation = 'pulse',
  count = 1,
}) => {
  const style: React.CSSProperties = {
    width,
    height: variant === 'text' ? '1em' : height,
  };

  const skeletonElement = (
    <div
      className={`skeleton skeleton-${variant} skeleton-${animation}`}
      style={style}
    />
  );

  if (count > 1) {
    return (
      <div className="skeleton-group">
        {Array.from({ length: count }).map((_, index) => (
          <React.Fragment key={index}>
            {skeletonElement}
          </React.Fragment>
        ))}
      </div>
    );
  }

  return skeletonElement;
};

// Pre-configured skeleton components
export const SkeletonCard: React.FC = () => (
  <div className="skeleton-card">
    <Skeleton variant="rectangular" width="100%" height={200} />
    <div className="skeleton-card-content">
      <Skeleton variant="text" width="60%" />
      <Skeleton variant="text" width="80%" />
      <Skeleton variant="text" width="40%" />
    </div>
  </div>
);

export const SkeletonTable: React.FC<{ rows?: number }> = ({ rows = 5 }) => (
  <div className="skeleton-table">
    {Array.from({ length: rows }).map((_, index) => (
      <div key={index} className="skeleton-table-row">
        <Skeleton variant="text" width="20%" />
        <Skeleton variant="text" width="30%" />
        <Skeleton variant="text" width="25%" />
        <Skeleton variant="text" width="15%" />
      </div>
    ))}
  </div>
);

export const SkeletonList: React.FC<{ items?: number }> = ({ items = 5 }) => (
  <div className="skeleton-list">
    {Array.from({ length: items }).map((_, index) => (
      <div key={index} className="skeleton-list-item">
        <Skeleton variant="circular" width={40} height={40} />
        <div className="skeleton-list-content">
          <Skeleton variant="text" width="70%" />
          <Skeleton variant="text" width="40%" />
        </div>
      </div>
    ))}
  </div>
);

export default Skeleton;
