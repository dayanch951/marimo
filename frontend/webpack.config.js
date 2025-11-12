const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const TerserPlugin = require('terser-webpack-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer');
const CompressionPlugin = require('compression-webpack-plugin');
const ImageMinimizerPlugin = require('image-minimizer-webpack-plugin');

const isProduction = process.env.NODE_ENV === 'production';
const enableAnalyzer = process.env.ANALYZE === 'true';

module.exports = {
  mode: isProduction ? 'production' : 'development',

  entry: {
    main: './src/index.tsx',
  },

  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: isProduction
      ? 'static/js/[name].[contenthash:8].js'
      : 'static/js/[name].js',
    chunkFilename: isProduction
      ? 'static/js/[name].[contenthash:8].chunk.js'
      : 'static/js/[name].chunk.js',
    publicPath: '/',
    clean: true,
    assetModuleFilename: 'static/media/[name].[hash][ext]',
  },

  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx', '.json'],
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@components': path.resolve(__dirname, 'src/components'),
      '@pages': path.resolve(__dirname, 'src/pages'),
      '@hooks': path.resolve(__dirname, 'src/hooks'),
      '@utils': path.resolve(__dirname, 'src/utils'),
      '@services': path.resolve(__dirname, 'src/services'),
      '@assets': path.resolve(__dirname, 'src/assets'),
    },
  },

  module: {
    rules: [
      // TypeScript and JavaScript
      {
        test: /\.(ts|tsx|js|jsx)$/,
        exclude: /node_modules/,
        use: {
          loader: 'babel-loader',
          options: {
            presets: [
              '@babel/preset-env',
              '@babel/preset-react',
              '@babel/preset-typescript',
            ],
            plugins: [
              // Enable lazy loading
              '@babel/plugin-syntax-dynamic-import',
              // Tree shaking
              isProduction && 'babel-plugin-transform-react-remove-prop-types',
              // Remove console.log in production
              isProduction && 'transform-remove-console',
            ].filter(Boolean),
          },
        },
      },

      // CSS
      {
        test: /\.css$/,
        use: [
          isProduction ? MiniCssExtractPlugin.loader : 'style-loader',
          {
            loader: 'css-loader',
            options: {
              modules: {
                auto: true,
                localIdentName: isProduction
                  ? '[hash:base64:8]'
                  : '[path][name]__[local]',
              },
              importLoaders: 1,
            },
          },
          'postcss-loader',
        ],
      },

      // Images
      {
        test: /\.(png|jpg|jpeg|gif|webp)$/i,
        type: 'asset',
        parser: {
          dataUrlCondition: {
            maxSize: 8 * 1024, // 8kb - inline smaller images
          },
        },
      },

      // SVG
      {
        test: /\.svg$/,
        use: ['@svgr/webpack', 'url-loader'],
      },

      // Fonts
      {
        test: /\.(woff|woff2|eot|ttf|otf)$/i,
        type: 'asset/resource',
        generator: {
          filename: 'static/fonts/[name].[hash][ext]',
        },
      },
    ],
  },

  plugins: [
    // HTML generation
    new HtmlWebpackPlugin({
      template: './public/index.html',
      inject: true,
      minify: isProduction && {
        removeComments: true,
        collapseWhitespace: true,
        removeRedundantAttributes: true,
        useShortDoctype: true,
        removeEmptyAttributes: true,
        removeStyleLinkTypeAttributes: true,
        keepClosingSlash: true,
        minifyJS: true,
        minifyCSS: true,
        minifyURLs: true,
      },
    }),

    // Extract CSS
    isProduction &&
      new MiniCssExtractPlugin({
        filename: 'static/css/[name].[contenthash:8].css',
        chunkFilename: 'static/css/[name].[contenthash:8].chunk.css',
      }),

    // Compression
    isProduction &&
      new CompressionPlugin({
        filename: '[path][base].gz',
        algorithm: 'gzip',
        test: /\.(js|css|html|svg)$/,
        threshold: 8192,
        minRatio: 0.8,
      }),

    // Brotli compression
    isProduction &&
      new CompressionPlugin({
        filename: '[path][base].br',
        algorithm: 'brotliCompress',
        test: /\.(js|css|html|svg)$/,
        compressionOptions: {
          level: 11,
        },
        threshold: 8192,
        minRatio: 0.8,
      }),

    // Bundle analyzer
    enableAnalyzer &&
      new BundleAnalyzerPlugin({
        analyzerMode: 'static',
        openAnalyzer: true,
        reportFilename: 'bundle-report.html',
      }),
  ].filter(Boolean),

  optimization: {
    minimize: isProduction,
    minimizer: [
      // Minify JavaScript
      new TerserPlugin({
        terserOptions: {
          parse: {
            ecma: 8,
          },
          compress: {
            ecma: 5,
            warnings: false,
            comparisons: false,
            inline: 2,
            drop_console: true,
          },
          mangle: {
            safari10: true,
          },
          output: {
            ecma: 5,
            comments: false,
            ascii_only: true,
          },
        },
        parallel: true,
      }),

      // Minify CSS
      new CssMinimizerPlugin(),

      // Optimize images
      isProduction &&
        new ImageMinimizerPlugin({
          minimizer: {
            implementation: ImageMinimizerPlugin.imageminMinify,
            options: {
              plugins: [
                ['imagemin-mozjpeg', { quality: 85 }],
                ['imagemin-pngquant', { quality: [0.8, 0.9] }],
                ['imagemin-svgo', {
                  plugins: [
                    {
                      name: 'removeViewBox',
                      active: false,
                    },
                  ],
                }],
              ],
            },
          },
          generator: [
            {
              preset: 'webp',
              implementation: ImageMinimizerPlugin.imageminGenerate,
              options: {
                plugins: ['imagemin-webp'],
              },
            },
          ],
        }),
    ].filter(Boolean),

    // Code splitting
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        // Vendor libraries
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          priority: 10,
          reuseExistingChunk: true,
        },

        // React libraries
        react: {
          test: /[\\/]node_modules[\\/](react|react-dom|react-router-dom)[\\/]/,
          name: 'react',
          priority: 20,
          reuseExistingChunk: true,
        },

        // UI libraries
        ui: {
          test: /[\\/]node_modules[\\/](@mui|@emotion)[\\/]/,
          name: 'ui',
          priority: 15,
          reuseExistingChunk: true,
        },

        // Common code
        common: {
          minChunks: 2,
          priority: 5,
          reuseExistingChunk: true,
          enforce: true,
        },
      },
    },

    // Runtime chunk
    runtimeChunk: {
      name: 'runtime',
    },

    // Module IDs
    moduleIds: 'deterministic',
  },

  // Performance hints
  performance: {
    hints: isProduction ? 'warning' : false,
    maxEntrypointSize: 250000, // 250KB
    maxAssetSize: 250000, // 250KB
  },

  // Dev server
  devServer: {
    port: 3000,
    hot: true,
    historyApiFallback: true,
    compress: true,
    open: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },

  // Source maps
  devtool: isProduction ? 'source-map' : 'eval-source-map',

  // Cache
  cache: {
    type: 'filesystem',
    cacheDirectory: path.resolve(__dirname, '.webpack-cache'),
    buildDependencies: {
      config: [__filename],
    },
  },

  // Stats
  stats: {
    preset: 'normal',
    moduleTrace: true,
    errorDetails: true,
  },
};
