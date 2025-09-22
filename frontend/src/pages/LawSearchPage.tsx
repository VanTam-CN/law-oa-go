import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  FaArrowLeft,
  FaMagnifyingGlass,
  FaBook,
  FaScaleBalanced,
  FaCalendar,
  FaTag,
  FaFileLines,
  FaEye,
  FaStar,
  FaChevronRight,
  FaTriangleExclamation
} from 'react-icons/fa6';

interface LawArticle {
  id: number;
  title: string;
  content: string;
  lawName: string;
  lawType: string;
  effectiveDate: string;
  category: string;
  isImportant: boolean;
  isFavorite: boolean;
}

const LawSearch: React.FC = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedType, setSelectedType] = useState<string>('all');
  const [results, setResults] = useState<LawArticle[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [showFavorites, setShowFavorites] = useState<boolean>(false);

  // 模拟法律条文数据
  const mockArticles: LawArticle[] = [
    {
      id: 1,
      title: '第一条 立法目的',
      content: '为了保护民事主体的合法权益，明确民事责任，维护社会和经济秩序，适应中国特色社会主义发展要求，根据宪法，制定本法。',
      lawName: '中华人民共和国民法典',
      lawType: '基本法律',
      effectiveDate: '2021-01-01',
      category: '总则',
      isImportant: true,
      isFavorite: false
    },
    {
      id: 2,
      title: '第一百二十七条 物权保护',
      content: '法律对数据、网络虚拟财产的保护有规定的，依照其规定。没有规定的，参照适用本法关于物权保护的规定。',
      lawName: '中华人民共和国民法典',
      lawType: '基本法律',
      effectiveDate: '2021-01-01',
      category: '物权编',
      isImportant: true,
      isFavorite: true
    },
    {
      id: 3,
      title: '第三条 平等原则',
      content: '民事主体在民事活动中的法律地位一律平等。',
      lawName: '中华人民共和国民法典',
      lawType: '基本法律',
      effectiveDate: '2021-01-01',
      category: '总则',
      isImportant: true,
      isFavorite: false
    },
    {
      id: 4,
      title: '第五百七十七条 违约责任',
      content: '当事人一方不履行合同义务或者履行合同义务不符合约定的，应当承担继续履行、采取补救措施或者赔偿损失等违约责任。',
      lawName: '中华人民共和国民法典',
      lawType: '基本法律',
      effectiveDate: '2021-01-01',
      category: '合同编',
      isImportant: true,
      isFavorite: false
    },
    {
      id: 5,
      title: '第十五条 自然人的民事权利能力',
      content: '自然人从出生时起到死亡时止，具有民事权利能力，依法享有民事权利，承担民事义务。',
      lawName: '中华人民共和国民法典',
      lawType: '基本法律',
      effectiveDate: '2021-01-01',
      category: '总则',
      isImportant: true,
      isFavorite: true
    }
  ];

  const categories = ['all', '总则', '物权编', '合同编', '人格权编', '婚姻家庭编', '继承编', '侵权责任编'];
  const lawTypes = ['all', '基本法律', '行政法规', '司法解释', '部门规章'];

  const handleSearch = async () => {
    if (!searchTerm.trim() && selectedCategory === 'all' && selectedType === 'all' && !showFavorites) {
      setError('请输入搜索关键词或选择筛选条件');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      let filteredArticles = mockArticles;

      // 应用筛选条件
      if (searchTerm.trim()) {
        filteredArticles = filteredArticles.filter(article =>
          article.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
          article.content.toLowerCase().includes(searchTerm.toLowerCase()) ||
          article.lawName.toLowerCase().includes(searchTerm.toLowerCase())
        );
      }

      if (selectedCategory !== 'all') {
        filteredArticles = filteredArticles.filter(article => article.category === selectedCategory);
      }

      if (selectedType !== 'all') {
        filteredArticles = filteredArticles.filter(article => article.lawType === selectedType);
      }

      if (showFavorites) {
        filteredArticles = filteredArticles.filter(article => article.isFavorite);
      }

      setResults(filteredArticles);
    } catch (error) {
      console.error('Search failed:', error);
      setError('搜索失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const toggleFavorite = (id: number) => {
    setResults(prev => prev.map(article =>
      article.id === id ? { ...article, isFavorite: !article.isFavorite } : article
    ));
  };

  const highlightSearchTerm = (text: string) => {
    if (!searchTerm.trim()) return text;

    const regex = new RegExp(`(${searchTerm})`, 'gi');
    return text.split(regex).map((part, index) =>
      regex.test(part) ? (
        <mark key={index} className="bg-warning">
          {part}
        </mark>
      ) : part
    );
  };

  useEffect(() => {
    // 如果有收藏筛选，自动搜索
    if (showFavorites) {
      handleSearch();
    }
  }, [showFavorites]);

  return (
    <div className="law-search p-4">
      <Card className="mb-4">
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <div className="d-flex align-items-center">
              <Button variant="outline-secondary" onClick={() => navigate('/tools')} className="me-3">
                <FaArrowLeft className="w-4 h-4" />
              </Button>
              <div>
                <h4 className="mb-0">法条查询</h4>
                <p className="text-muted mb-0">快速查询相关法律条文</p>
              </div>
            </div>
            <Badge bg="primary">
              <FaBook className="w-4 h-4 me-1" />
              工具
            </Badge>
          </div>
        </Card.Header>
        <Card.Body>
          <Alert variant="info">
            <span className="me-2">ℹ️</span>
            <strong>使用说明：</strong>
            输入关键词搜索法律条文，可按分类、法律类型筛选，也可查看收藏的法条。
          </Alert>

          {/* 搜索表单 */}
          <Form>
            <Form.Group className="mb-3">
              <Form.Label>搜索关键词</Form.Label>
              <InputGroup>
                <InputGroup.Text>
                  <FaMagnifyingGlass className="w-4 h-4" />
                </InputGroup.Text>
                <Form.Control
                  type="text"
                  placeholder="输入关键词搜索法条标题或内容..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                />
              </InputGroup>
            </Form.Group>

            <div className="row mb-3">
              <div className="col-md-4">
                <Form.Label>法律分类</Form.Label>
                <Form.Select
                  value={selectedCategory}
                  onChange={(e) => setSelectedCategory(e.target.value)}
                >
                  <option value="all">全部分类</option>
                  {categories.filter(cat => cat !== 'all').map(category => (
                    <option key={category} value={category}>{category}</option>
                  ))}
                </Form.Select>
              </div>
              <div className="col-md-4">
                <Form.Label>法律类型</Form.Label>
                <Form.Select
                  value={selectedType}
                  onChange={(e) => setSelectedType(e.target.value)}
                >
                  <option value="all">全部类型</option>
                  {lawTypes.filter(type => type !== 'all').map(type => (
                    <option key={type} value={type}>{type}</option>
                  ))}
                </Form.Select>
              </div>
              <div className="col-md-4">
                <Form.Label>筛选条件</Form.Label>
                <Button
                  variant={showFavorites ? "warning" : "outline-warning"}
                  onClick={() => setShowFavorites(!showFavorites)}
                  className="w-100"
                >
                  <FaStar className={`w-4 h-4 me-2 ${showFavorites ? 'text-white' : ''}`} />
                  {showFavorites ? '显示全部' : '仅显示收藏'}
                </Button>
              </div>
            </div>

            <Button
              variant="primary"
              onClick={handleSearch}
              disabled={loading}
            >
              {loading ? (
                <>
                  <span className="spinner-border spinner-border-sm me-2" role="status" />
                  搜索中...
                </>
              ) : (
                <>
                  <MagnifyingGlassIcon className="w-4 h-4 me-2" />
                  搜索
                </>
              )}
            </Button>
          </Form>
        </Card.Body>
      </Card>

      {/* 搜索结果 */}
      {error && (
        <Alert variant="danger" onClose={() => setError(null)} dismissible>
          {error}
        </Alert>
      )}

      {results.length > 0 && (
        <Card>
          <Card.Header>
            <div className="d-flex justify-content-between align-items-center">
              <h5 className="mb-0">搜索结果</h5>
              <Badge bg="success">
                找到 {results.length} 条相关法条
              </Badge>
            </div>
          </Card.Header>
          <Card.Body>
            <div className="g">
              {results.map((article) => (
                <Card key={article.id} className="border-0 shadow-sm">
                  <Card.Body>
                    <div className="d-flex justify-content-between align-items-start mb-2">
                      <div className="flex-grow-1">
                        <h6 className="mb-1">
                          {highlightSearchTerm(article.title)}
                          {article.isImportant && (
                            <Badge bg="danger" className="ms-2">重要</Badge>
                          )}
                        </h6>
                        <div className="d-flex gap-3 mb-2">
                          <small className="text-muted">
                            <FaScaleBalanced className="w-3 h-3 me-1" />
                            {article.lawName}
                          </small>
                          <small className="text-muted">
                            <FaTag className="w-3 h-3 me-1" />
                            {article.category}
                          </small>
                          <small className="text-muted">
                            <FaCalendar className="w-3 h-3 me-1" />
                            {article.effectiveDate}
                          </small>
                        </div>
                      </div>
                      <div className="d-flex gap-2">
                        <Button
                          variant="outline-warning"
                          size="sm"
                          onClick={() => toggleFavorite(article.id)}
                        >
                          <FaStar className={`w-4 h-4 ${article.isFavorite ? 'text-warning' : ''}`} />
                        </Button>
                        <Button
                          variant="outline-primary"
                          size="sm"
                          onClick={() => alert(`查看法条详情：${article.id}`)}
                        >
                          <FaEye className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>

                    <Card.Text className="text-muted">
                      {highlightSearchTerm(article.content)}
                    </Card.Text>

                    <div className="d-flex justify-content-between align-items-center">
                      <Badge bg="light" text="dark">
                        {article.lawType}
                      </Badge>
                      <Button
                        variant="outline-primary"
                        size="sm"
                        onClick={() => alert(`查看相关案例和司法解释`)}
                      >
                        <FaChevronRight className="w-4 h-4 me-1" />
                        查看相关
                      </Button>
                    </div>
                  </Card.Body>
                </Card>
              ))}
            </div>
          </Card.Body>
        </Card>
      )}

      {results.length === 0 && !loading && !error && searchTerm && (
        <Card>
          <Card.Body className="text-center">
            <FaFileLines className="w-16 h-16 text-muted mx-auto mb-3" />
            <h5>未找到相关法条</h5>
            <p className="text-muted">请尝试调整搜索关键词或筛选条件</p>
          </Card.Body>
        </Card>
      )}
    </div>
  );
};

export default LawSearch;