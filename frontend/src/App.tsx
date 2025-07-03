import React from 'react';
import { ApolloProvider } from '@apollo/client';
import { Layout, Typography } from 'antd';
import { ThunderboltOutlined } from '@ant-design/icons';
import { client } from './lib/apollo';
import { TorrentList } from './components/TorrentList';
import './App.css';

const { Header, Content } = Layout;
const { Title } = Typography;

function App() {
  return (
    <ApolloProvider client={client}>
      <Layout style={{ minHeight: '100vh', background: '#fafafa' }}>
        <Header
          style={{
            background: 'white',
            boxShadow: '0 2px 20px rgba(0, 0, 0, 0.08)',
            display: 'flex',
            alignItems: 'center',
            padding: '0 40px',
            position: 'sticky',
            top: 0,
            zIndex: 100,
            height: 70,
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{
              width: 40,
              height: 40,
              background: '#1890ff',
              borderRadius: 12,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              boxShadow: '0 4px 20px rgba(24, 144, 255, 0.3)'
            }}>
              <ThunderboltOutlined style={{ fontSize: 20, color: 'white' }} />
            </div>
            <Title level={3} style={{
              color: '#262626',
              margin: 0,
              fontWeight: 600,
              letterSpacing: '-0.5px'
            }}>
              SkTorrent
            </Title>
          </div>
        </Header>

        <Content style={{
          padding: '40px 5vw',
          background: '#fafafa',
          minHeight: 'calc(100vh - 70px)'
        }}>
          <TorrentList pageSize={24} />
        </Content>
      </Layout>
    </ApolloProvider>
  );
}

export default App;
