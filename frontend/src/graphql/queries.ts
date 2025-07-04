import { gql } from '@apollo/client';

export const GET_TORRENTS = gql`
  query GetTorrents(
    $first: Int
    $after: String
    $category: String
    $search: String
    $sortBy: TorrentSortBy
  ) {
    torrents(
      first: $first
      after: $after
      category: $category
      search: $search
      sortBy: $sortBy
    ) {
      torrents {
        id
        name
        category
        sizeMB
        addedDate
        url
        imageURL
        csfdRating
        csfdURL
        createdAt
        updatedAt
        seeds
        leeches
      }
      totalCount
      hasNextPage
      hasPreviousPage
    }
  }
`;

export const GET_RECENT_TORRENTS = gql`
  query GetRecentTorrents($limit: Int) {
    recentTorrents(limit: $limit) {
      id
      name
      category
      sizeMB
      addedDate
      url
      imageURL
      csfdRating
      csfdURL
      createdAt
      updatedAt
      seeds
      leeches
    }
  }
`;

export const SEARCH_TORRENTS = gql`
  query SearchTorrents($query: String!, $limit: Int) {
    searchTorrents(query: $query, limit: $limit) {
      id
      name
      category
      sizeMB
      addedDate
      url
      imageURL
      csfdRating
      csfdURL
      createdAt
      updatedAt
      seeds
      leeches
    }
  }
`;

export const GET_CATEGORIES = gql`
  query GetCategories {
    categories {
      name
      count
    }
  }
`;

export const GET_STATS = gql`
  query GetStats {
    stats {
      totalTorrents
      totalCategories
      categoryCounts {
        name
        count
      }
    }
  }
`;

export const GET_TORRENTS_BY_CSFD_ID = gql`
  query GetTorrentsByCSFDID($csfdID: String!, $limit: Int) {
    torrentsByCSFDID(csfdID: $csfdID, limit: $limit) {
      id
      name
      category
      sizeMB
      addedDate
      url
      imageURL
      csfdRating
      csfdURL
      createdAt
      updatedAt
      seeds
      leeches
    }
  }
`;
