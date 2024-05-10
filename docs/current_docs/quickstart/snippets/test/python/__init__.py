import dagger
from dagger import dag, function, object_type


@object_type
class HelloDagger:
    @function
    async def test(self, source: dagger.Directory) -> str:
        """Returns the result of running unit tests"""
        return await (
            dag.container()
            .from_("node:21-slim")
            .with_directory("/src", source.without_directory("dagger"))
            .with_workdir("/src")
            .with_exec(["npm", "install"])
            .with_exec(["npm", "run", "test:unit", "run"])
            .stdout()
        )