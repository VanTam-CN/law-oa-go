//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// LegalStatute 法条结构体
type LegalStatute struct {
	ID                 int       `json:"id" gorm:"primaryKey"`
	StatuteNumber      string    `json:"statute_number" gorm:"uniqueIndex;not null"`
	Title              string    `json:"title" gorm:"not null"`
	Content            string    `json:"content" gorm:"not null"`
	CategoryID         int       `json:"category_id"`
	LawName            string    `json:"law_name" gorm:"not null"`
	Chapter            string    `json:"chapter"`
	Section            string    `json:"section"`
	Part               string    `json:"part"`
	EffectiveDate      *time.Time `json:"effective_date"`
	ExpiryDate         *time.Time `json:"expiry_date"`
	PublishingAuthority string    `json:"publishing_authority"`
	Status             string    `json:"status" gorm:"default:active"`
	HierarchyLevel     int       `json:"hierarchy_level" gorm:"default:1"`
	ParentStatuteID    *int      `json:"parent_statute_id"`
	OrderInHierarchy   *int      `json:"order_in_hierarchy"`
	Tags               pq.StringArray `json:"tags" gorm:"type:text[]"`
	Keywords           pq.StringArray `json:"keywords" gorm:"type:text[]"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// LegalCategory 法条分类
type LegalCategory struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null"`
	ParentID    *int      `json:"parent_id"`
	Level       int       `json:"level" gorm:"default:1"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file")
	}

	// 数据库连接配置
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "law_oa_user")
	dbPassword := getEnv("DB_PASSWORD", "your_secure_password")
	dbName := getEnv("DB_NAME", "law_oa_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}
	defer sqlDB.Close()

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	log.Println("数据库连接成功")

	// 检查是否已经存在法条数据
	var count int64
	db.Model(&LegalStatute{}).Count(&count)
	if count > 0 {
		log.Printf("法条数据已存在，跳过数据迁移。当前法条数量: %d", count)

		// 即使数据已存在，也要生成ES数据文件
		log.Println("生成Elasticsearch索引数据文件...")
		var existingStatutes []LegalStatute
		db.Find(&existingStatutes)
		if err := generateESDataFile(existingStatutes); err != nil {
			log.Printf("生成ES数据文件失败: %v", err)
		} else {
			log.Println("ES数据文件生成完成: legal_statutes_es_data.json")
		}
		return
	}

	// 准备示例数据
	statutes := prepareSampleStatutes()
	categories := prepareSampleCategories()

	// 导入分类数据
	log.Println("开始导入法条分类数据...")
	if err := importCategories(db, categories); err != nil {
		log.Fatalf("导入分类数据失败: %v", err)
	}

	// 导入法条数据
	log.Println("开始导入法条数据...")
	if err := importStatutes(db, statutes); err != nil {
		log.Fatalf("导入法条数据失败: %v", err)
	}

	log.Printf("法条数据迁移完成！共导入 %d 个分类，%d 个法条", len(categories), len(statutes))

	// 生成Elasticsearch索引数据
	log.Println("生成Elasticsearch索引数据文件...")
	if err := generateESDataFile(statutes); err != nil {
		log.Printf("生成ES数据文件失败: %v", err)
	} else {
		log.Println("ES数据文件生成完成: legal_statutes_es_data.json")
	}
}

func prepareSampleCategories() []LegalCategory {
	return []LegalCategory{
		{
			Name:        "民法",
			Code:        "CIVIL_LAW",
			Level:       1,
			Description: "民法相关法律法规",
			IsActive:    true,
		},
		{
			Name:        "刑法",
			Code:        "CRIMINAL_LAW",
			Level:       1,
			Description: "刑法相关法律法规",
			IsActive:    true,
		},
		{
			Name:        "商法",
			Code:        "COMMERCIAL_LAW",
			Level:       1,
			Description: "商法相关法律法规",
			IsActive:    true,
		},
		{
			Name:        "劳动法",
			Code:        "LABOR_LAW",
			Level:       1,
			Description: "劳动法相关法律法规",
			IsActive:    true,
		},
		{
			Name:        "公司法",
			Code:        "COMPANY_LAW",
			Level:       2,
			Description: "公司法相关法律法规",
			IsActive:    true,
		},
	}
}

func prepareSampleStatutes() []LegalStatute {
	now := time.Now()
	statutes := []LegalStatute{
		// 民法典相关法条
		{
			StatuteNumber: "民法典第一千一百九十二条",
			Title:         "个人劳务关系中的侵权责任",
			Content:       "个人之间形成劳务关系，提供劳务一方因劳务造成他人损害的，由接受劳务一方承担侵权责任。接受劳务一方承担侵权责任后，可以向有故意或者重大过失的提供劳务一方追偿。提供劳务一方因劳务受到损害的，根据双方各自的过错承担相应的责任。",
			LawName:       "中华人民共和国民法典",
			Chapter:       "第七章 侵权责任",
			Section:       "第一节 损害赔偿",
			EffectiveDate: &now,
			Status:        "active",
			HierarchyLevel: 3,
			Tags:          []string{"常用", "基础"},
			Keywords:      []string{"侵权责任", "劳务关系", "损害赔偿"},
		},
		{
			StatuteNumber: "民法典第一千二百零六条",
			Title:         "产品责任",
			Content:       "产品投入流通后发现存在缺陷的，生产者、销售者应当及时采取警示、召回等补救措施。未及时采取补救措施或者补救措施不力造成损害扩大的，对扩大的损害也应当承担侵权责任。",
			LawName:       "中华人民共和国民法典",
			Chapter:       "第七章 侵权责任",
			Section:       "第四节 产品责任",
			EffectiveDate: &now,
			Status:        "active",
			HierarchyLevel: 3,
			Tags:          []string{"重要", "常用"},
			Keywords:      []string{"产品责任", "缺陷产品", "召回", "警示"},
		},
		// 公司法相关法条
		{
			StatuteNumber: "公司法第三条",
			Title:         "公司定义及法律责任",
			Content:       "公司是企业法人，有独立的法人财产，享有法人财产权。公司以其全部财产对公司的债务承担责任。有限责任公司的股东以其认缴的出资额为限对公司承担责任；股份有限公司的股东以其认购的股份为限对公司承担责任。",
			LawName:       "中华人民共和国公司法",
			Chapter:       "第一章 总则",
			EffectiveDate: &now,
			Status:        "active",
			HierarchyLevel: 2,
			Tags:          []string{"基础", "重要"},
			Keywords:      []string{"公司", "法人财产", "有限责任", "股份责任"},
		},
		{
			StatuteNumber: "公司法第四条",
			Title:         "股东权利",
			Content:       "公司股东依法享有资产收益、参与重大决策和选择管理者等权利。公司股东应当遵守法律、行政法规和公司章程，依法行使股东权利，不得滥用股东权利损害公司或者其他股东的利益。",
			LawName:       "中华人民共和国公司法",
			Chapter:       "第一章 总则",
			EffectiveDate: &now,
			Status:        "active",
			HierarchyLevel: 2,
			Tags:          []string{"基础", "重要"},
			Keywords:      []string{"股东权利", "资产收益", "重大决策", "管理者选择"},
		},
		// 劳动法相关法条
		{
			StatuteNumber: "劳动合同法第三十九条",
			Title:         "用人单位单方解除劳动合同",
			Content:       "劳动者有下列情形之一的，用人单位可以解除劳动合同：（一）在试用期间被证明不符合录用条件的；（二）严重违反用人单位的规章制度的；（三）严重失职，营私舞弊，给用人单位造成重大损害的；（四）劳动者同时与其他用人单位建立劳动关系，对完成本单位的工作任务造成严重影响，或者经用人单位提出，拒不改正的；（五）因欺诈、胁迫等手段使对方在违背真实意思的情况下订立或者变更劳动合同，致使劳动合同无效的。",
			LawName:       "中华人民共和国劳动合同法",
			Chapter:       "第四章 解除和终止劳动合同",
			EffectiveDate: &now,
			Status:        "active",
			HierarchyLevel: 2,
			Tags:          []string{"常用", "重要"},
			Keywords:      []string{"解除劳动合同", "试用期", "规章制度", "重大损害"},
		},
		{
			StatuteNumber: "劳动合同法第四十七条",
			Title:         "经济补偿的计算",
			Content:       "经济补偿按劳动者在本单位工作的年限，每满一年支付一个月工资的标准向劳动者支付。六个月以上不满一年的，按一年计算；不满六个月的，向劳动者支付半个月工资的经济补偿。",
			LawName:       "中华人民共和国劳动合同法",
			Chapter:       "第五章 特别规定",
			Section:       "第一节 集体合同",
			EffectiveDate: &now,
			Status:        "active",
			HierarchyLevel: 2,
			Tags:          []string{"常用", "基础"},
			Keywords:      []string{"经济补偿", "工作年限", "工资标准"},
		},
	}

	// 设置分类ID
	for i := range statutes {
		switch statutes[i].LawName {
		case "中华人民共和国民法典":
			statutes[i].CategoryID = 1 // 民法
		case "中华人民共和国公司法":
			statutes[i].CategoryID = 3 // 商法下的公司法
		case "中华人民共和国劳动合同法":
			statutes[i].CategoryID = 4 // 劳动法
		}
	}

	return statutes
}

func importCategories(db *gorm.DB, categories []LegalCategory) error {
	for _, category := range categories {
		// 检查是否已存在
		var existingCategory LegalCategory
		result := db.Where("code = ?", category.Code).First(&existingCategory)
		if result.Error == nil {
			log.Printf("分类 %s 已存在，跳过", category.Name)
			continue
		}

		if err := db.Create(&category).Error; err != nil {
			log.Printf("导入分类 %s 失败: %v", category.Name, err)
			continue
		}
		log.Printf("导入分类: %s (%s)", category.Name, category.Code)
	}
	return nil
}

func importStatutes(db *gorm.DB, statutes []LegalStatute) error {
	for _, statute := range statutes {
		// 检查是否已存在
		var existingStatute LegalStatute
		result := db.Where("statute_number = ?", statute.StatuteNumber).First(&existingStatute)
		if result.Error == nil {
			log.Printf("法条 %s 已存在，跳过", statute.StatuteNumber)
			continue
		}

		if err := db.Create(&statute).Error; err != nil {
			log.Printf("导入法条 %s 失败: %v", statute.StatuteNumber, err)
			continue
		}
		log.Printf("导入法条: %s - %s", statute.StatuteNumber, statute.Title)
	}
	return nil
}

func generateESDataFile(statutes []LegalStatute) error {
	// 获取分类信息映射
	categoryMap := map[int]string{
		1: "民法",
		2: "刑法",
		3: "商法",
		4: "劳动法",
	}

	esDocuments := make([]map[string]interface{}, 0, len(statutes))

	for _, statute := range statutes {
		doc := map[string]interface{}{
			"id":                    statute.ID,
			"statute_number":        statute.StatuteNumber,
			"title":                 statute.Title,
			"content":               statute.Content,
			"law_name":              statute.LawName,
			"chapter":               statute.Chapter,
			"section":               statute.Section,
			"category": map[string]interface{}{
				"id":   statute.CategoryID,
				"name": categoryMap[statute.CategoryID],
				"code": getCategoryCode(statute.CategoryID),
			},
			"effective_date":        statute.EffectiveDate.Format("2006-01-02"),
			"status":                statute.Status,
			"hierarchy_level":       statute.HierarchyLevel,
			"tags":                  statute.Tags,
			"keywords":              statute.Keywords,
			"created_at":            statute.CreatedAt.Format(time.RFC3339),
			"updated_at":            statute.UpdatedAt.Format(time.RFC3339),
			"content_length":        len(statute.Content),
			"view_count":            0,
			"favorite_count":        0,
			"search_weight":         1.0,
		}

		esDocuments = append(esDocuments, doc)
	}

	// 写入JSON文件
	file, err := os.Create("legal_statutes_es_data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(esDocuments); err != nil {
		return err
	}

	return nil
}

func getCategoryCode(categoryID int) string {
	codes := map[int]string{
		1: "CIVIL_LAW",
		2: "CRIMINAL_LAW",
		3: "COMMERCIAL_LAW",
		4: "LABOR_LAW",
	}
	return codes[categoryID]
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}