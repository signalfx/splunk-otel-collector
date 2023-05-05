var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

app.MapGet("/api/values/{id}", (int id) =>
{
    return $"value{id}";
});

app.Run();
