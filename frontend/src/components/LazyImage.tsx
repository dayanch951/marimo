import React, { useState, useEffect, useRef } from 'react';

interface LazyImageProps {
  src: string;
  alt: string;
  placeholder?: string;
  className?: string;
  onLoad?: () => void;
  onError?: () => void;
  threshold?: number;
  srcSet?: string;
  sizes?: string;
}

/**
 * LazyImage component with Intersection Observer for lazy loading
 */
export const LazyImage: React.FC<LazyImageProps> = ({
  src,
  alt,
  placeholder,
  className = '',
  onLoad,
  onError,
  threshold = 0.1,
  srcSet,
  sizes,
}) => {
  const [imageSrc, setImageSrc] = useState<string>(placeholder || '');
  const [imageRef, setImageRef] = useState<HTMLImageElement | null>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInView, setIsInView] = useState(false);

  useEffect(() => {
    if (!imageRef || !('IntersectionObserver' in window)) {
      // Fallback: load immediately if no IntersectionObserver support
      setImageSrc(src);
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setIsInView(true);
            observer.unobserve(entry.target);
          }
        });
      },
      {
        threshold,
        rootMargin: '50px', // Start loading 50px before entering viewport
      }
    );

    observer.observe(imageRef);

    return () => {
      if (imageRef) {
        observer.unobserve(imageRef);
      }
    };
  }, [imageRef, src, threshold]);

  useEffect(() => {
    if (isInView && !isLoaded) {
      // Preload image
      const img = new Image();
      img.src = src;
      if (srcSet) img.srcset = srcSet;

      img.onload = () => {
        setImageSrc(src);
        setIsLoaded(true);
        onLoad?.();
      };

      img.onerror = () => {
        onError?.();
      };
    }
  }, [isInView, src, srcSet, isLoaded, onLoad, onError]);

  return (
    <img
      ref={setImageRef}
      src={imageSrc}
      srcSet={isLoaded ? srcSet : undefined}
      sizes={sizes}
      alt={alt}
      className={`${className} ${isLoaded ? 'loaded' : 'loading'}`}
      style={{
        filter: isLoaded ? 'none' : 'blur(10px)',
        transition: 'filter 0.3s ease-in-out',
      }}
    />
  );
};

/**
 * Progressive Image component with low-quality placeholder
 */
interface ProgressiveImageProps {
  src: string;
  placeholderSrc: string;
  alt: string;
  className?: string;
}

export const ProgressiveImage: React.FC<ProgressiveImageProps> = ({
  src,
  placeholderSrc,
  alt,
  className = '',
}) => {
  const [currentSrc, setCurrentSrc] = useState(placeholderSrc);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const img = new Image();
    img.src = src;

    img.onload = () => {
      setCurrentSrc(src);
      setIsLoaded(true);
    };
  }, [src]);

  return (
    <div className={`progressive-image ${className}`}>
      <img
        src={currentSrc}
        alt={alt}
        style={{
          filter: isLoaded ? 'none' : 'blur(20px)',
          transform: isLoaded ? 'scale(1)' : 'scale(1.1)',
          transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
        }}
      />
    </div>
  );
};

/**
 * Responsive Image component with srcset support
 */
interface ResponsiveImageProps {
  src: string;
  srcSet: string;
  sizes: string;
  alt: string;
  className?: string;
  lazy?: boolean;
}

export const ResponsiveImage: React.FC<ResponsiveImageProps> = ({
  src,
  srcSet,
  sizes,
  alt,
  className = '',
  lazy = true,
}) => {
  if (lazy) {
    return (
      <LazyImage
        src={src}
        srcSet={srcSet}
        sizes={sizes}
        alt={alt}
        className={className}
      />
    );
  }

  return (
    <img
      src={src}
      srcSet={srcSet}
      sizes={sizes}
      alt={alt}
      className={className}
      loading="lazy"
    />
  );
};

/**
 * Background Image component with lazy loading
 */
interface LazyBackgroundProps {
  imageUrl: string;
  className?: string;
  children?: React.ReactNode;
}

export const LazyBackground: React.FC<LazyBackgroundProps> = ({
  imageUrl,
  className = '',
  children,
}) => {
  const [isLoaded, setIsLoaded] = useState(false);
  const [backgroundImage, setBackgroundImage] = useState<string>('none');
  const elementRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!elementRef.current || !('IntersectionObserver' in window)) {
      setBackgroundImage(`url(${imageUrl})`);
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const img = new Image();
            img.src = imageUrl;

            img.onload = () => {
              setBackgroundImage(`url(${imageUrl})`);
              setIsLoaded(true);
            };

            observer.unobserve(entry.target);
          }
        });
      },
      {
        threshold: 0.1,
        rootMargin: '50px',
      }
    );

    observer.observe(elementRef.current);

    return () => {
      if (elementRef.current) {
        observer.unobserve(elementRef.current);
      }
    };
  }, [imageUrl]);

  return (
    <div
      ref={elementRef}
      className={`${className} ${isLoaded ? 'loaded' : 'loading'}`}
      style={{
        backgroundImage,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
        transition: 'opacity 0.3s ease-in-out',
        opacity: isLoaded ? 1 : 0.5,
      }}
    >
      {children}
    </div>
  );
};

/**
 * useImagePreload hook for preloading images
 */
export const useImagePreload = (urls: string[]): boolean => {
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    if (urls.length === 0) {
      setIsLoaded(true);
      return;
    }

    let loadedCount = 0;
    const totalImages = urls.length;

    urls.forEach((url) => {
      const img = new Image();
      img.src = url;

      img.onload = () => {
        loadedCount++;
        if (loadedCount === totalImages) {
          setIsLoaded(true);
        }
      };

      img.onerror = () => {
        loadedCount++;
        if (loadedCount === totalImages) {
          setIsLoaded(true);
        }
      };
    });
  }, [urls]);

  return isLoaded;
};

/**
 * Image optimization utilities
 */
export const getOptimizedImageUrl = (
  baseUrl: string,
  options: {
    width?: number;
    height?: number;
    quality?: number;
    format?: 'webp' | 'jpeg' | 'png';
  } = {}
): string => {
  const params = new URLSearchParams();

  if (options.width) params.append('w', options.width.toString());
  if (options.height) params.append('h', options.height.toString());
  if (options.quality) params.append('q', options.quality.toString());
  if (options.format) params.append('f', options.format);

  const queryString = params.toString();
  return queryString ? `${baseUrl}?${queryString}` : baseUrl;
};

/**
 * Generate srcset for responsive images
 */
export const generateSrcSet = (
  baseUrl: string,
  widths: number[],
  format: 'webp' | 'jpeg' = 'webp'
): string => {
  return widths
    .map((width) => {
      const url = getOptimizedImageUrl(baseUrl, { width, format, quality: 85 });
      return `${url} ${width}w`;
    })
    .join(', ');
};

/**
 * Generate sizes attribute for responsive images
 */
export const generateSizes = (
  breakpoints: { maxWidth: number; size: string }[]
): string => {
  return breakpoints
    .map((bp) => `(max-width: ${bp.maxWidth}px) ${bp.size}`)
    .join(', ');
};
