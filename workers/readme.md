# Temporal Workers

## [How to create custom search attribnute?](https://docs.temporal.io/self-hosted-guide/visibility#create-custom-search-attributes)



```bash
temporal operator search-attribute create --name="OrderProcessingStatus" --type="Keyword"
```

## How to register namespace?

```bash
tctl --ns oms-dev namespace register -rd 1
```