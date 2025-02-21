# Models

In GRF, models are employed to specify the schema of the data that you want to read or write. For performance reasons, the data is internally handled as maps and only converted to model structs when required (e.g., in the GORM query driver during create and update operations). The `models.InternalValue` type serves as a map representation of the model across the framework.


## Introduction

### Requirements

:::info
To use GRF, you need to have a single primary key (either an `u?int[0-9]{2}` or a string) that is annotated with the `id` JSON tag. It's not possible to use composite primary keys, and you also can't use GRF without a primary key. 
:::

GRF only mandates the struct to have an exported field with an `id` JSON tag. However, depending on the Query Driver implementation, additional requirements may be necessary. For instance, if you intend to store your models in a SQL database (e.g., with the default GORM Query Driver), you must ensure that all fields implement the sql.Scanner and driver.Valuer interfaces.

```go
// this will work
type ValidModel struct {
    ID string `json:"id"`
}

// this will fail
type InvalidModel{
    Foo string
}
```

The `id` field requirement is non-negotiable, as it is used to uniquely identify the model in the storage. The `id` field can be either a numeric (assumed auto incremented ID field) or a string (assumed UUID, but any will be fine, as long as it's unique) type.


### Conversion between struct and `models.InternalValue`

GRF exposes two functions in `models` package, that allow conversion between struct and `models.InternalValue`:

* `models.AsInternalValue` - converts a struct to `models.InternalValue`
* `models.AsModel` - converts `models.InternalValue` to a struct

The functions use reflection to convert between the types, so they are not the fastest. However, they are very convenient, and you can always implement your own conversion functions, if the speed is an issue.


## Model fields

All the model fields must be tagged with a `json` tag, which is used to map the field name in `models.InternalValue`, and thus to (at least default) request and response JSON payloads. Saying that, GRF must know how to:

* Read the field from the storage (in case of GORM, the field should implement the sql.Scanner interface)
* Write the field to the storage (in case of GORM, the field should implement the driver.Valuer interface)
* Read the field from the request JSON payload (GRF uses some tricks to do that, read more in the [Serializers](./serializers) section)
* Write the field to the response JSON payload, read more in the [Serializers](./serializers) section, like above.


:::warning
Those types are not supported yet, but will be in the future:
* `slice<int>`
* pointer fields, for example `*string`
* JSON fields, as non-string, dedicated column types (eg. Postgres)
:::

### Slice field

`models.SliceField` can be used to store slices, that are encoded to JSON string for storage (implement sql.Scanner and driver.Valuer interfaces). In request and response JSON payloads, the slice is represented as a JSON array. The types of the slice need to be golang built-in basic types. The field provides validation of all the elements in the slice.

### JSON fields

`datatypes.JSON` from `gorm.datatypes` package can be used to store JSON data in a database, in JSON column type native to the database. The field is represented as a JSON object in the request and response JSON payloads.

## Model relations

GRF models by themselves do not directly support relations, but:

* grf allows setting a `grf:"relation"` tag on the field, that instructs serializers to not treat a field as a basic type and skip initial parsing
* GORM's query driver `WithPreload` method can be used to [preload](https://gorm.io/docs/preload.html) related models
* `fields.SerializerField` can be used to include related models in the response JSON payload as a nested object

All of this together allows for a simple implementation of relations in GRF:

```go

type Profile struct {
	models.BaseModel
	Name   string  `json:"name" gorm:"size:191;column:name"`
	Photos []Photo `json:"photos" gorm:"foreignKey:profile_id" grf:"relation"`
}

type Photo struct {
	models.BaseModel
	ProfileID uuid.UUID `json:"profile_id" gorm:"size:191;column:profile_id"`
}

views.NewModelViewSet[Profile](
		"/profiles",
		queries.GORM[Profile](
			gormDB,
		).WithPreload(
			"photos",  // Note JSON tag here, in original GORM API it's the field name
		).WithOrderBy(
			"`profiles`.`created_at` ASC",
		),
	).WithSerializer(
		serializers.NewModelSerializer[Profile]().WithNewField(
			serializers.NewSerializerField[Photo](
				"photos",
				serializers.NewModelSerializer[Photo](),
			),
		),
	).Register(router)
```

:::warning
    GORM's Joins are not supported, as they are pretty useless anyway. If you need to join tables, you have no choice but to create a view in your SQL database and use it as a model.
:::
