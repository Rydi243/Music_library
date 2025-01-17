// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/song": {
            "get": {
                "description": "Возвращает список песен с пагинацией и фильтрацией по всем полям",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Получить список песен",
                "parameters": [
                    {
                        "description": "Параметры",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/structSong.ReqGetSong"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Список песен",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/structSong.Song"
                            }
                        }
                    },
                    "400": {
                        "description": "Ошибка декодирования",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка запроса к базе данных",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "put": {
                "description": "Обновляет информацию о песне",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Обновить данные песни",
                "parameters": [
                    {
                        "description": "Информация для обновления",
                        "name": "song",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/structSong.Song"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Песня обновлена",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Ошибка декодирования или отсутствуют данные для обновления",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Песня не найдена",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка обновления записи в базе данных",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "post": {
                "description": "Добавляет новую песню с тестового апи",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Добавить песню(обращается к тестовому апи)",
                "parameters": [
                    {
                        "description": "Данные песни",
                        "name": "song",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/structSong.Song"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Песня добавлена",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Ошибка декодирования",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "409": {
                        "description": "Песня уже существует",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка при получении данных о песне или вставке в базу данных",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "description": "Удаляет песню из библиотеки по её названию",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Удалить песню",
                "parameters": [
                    {
                        "description": "Название песни для удаления",
                        "name": "song",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/structSong.Song"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Песня удалена",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Ошибка декодирования или отсутствует название песни",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Песня не найдена",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка при удалении записи из базы данных",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/song/text": {
            "post": {
                "description": "Возвращает текст песни, разделенный на куплеты, с пагинацией",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Получить текст песни с пагинацией по куплетам",
                "parameters": [
                    {
                        "description": "Параметры",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/structSong.ReqTextSong"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Список куплетов",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Ошибка декодирования или отсутствует название песни",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Песня не найдена",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка при выполнении запроса к базе данных",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "structSong.ReqGetSong": {
            "type": "object",
            "properties": {
                "group": {
                    "type": "string"
                },
                "limit": {
                    "type": "integer"
                },
                "link": {
                    "type": "string"
                },
                "offset": {
                    "type": "integer"
                },
                "release_date": {
                    "type": "string"
                },
                "song": {
                    "type": "string"
                },
                "text": {
                    "type": "string"
                }
            }
        },
        "structSong.ReqTextSong": {
            "type": "object",
            "properties": {
                "limit": {
                    "type": "integer"
                },
                "offset": {
                    "type": "integer"
                },
                "song": {
                    "type": "string"
                }
            }
        },
        "structSong.Song": {
            "type": "object",
            "properties": {
                "group": {
                    "type": "string"
                },
                "link": {
                    "type": "string"
                },
                "release_date": {
                    "type": "string"
                },
                "song": {
                    "type": "string"
                },
                "text": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Music_library",
	Description:      "Добавляет, запрашивает песни с пагинацией. запрашивает песни с пагинацией по куплетам. Обновляет и удаляет песни.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
